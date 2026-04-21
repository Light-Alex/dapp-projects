package nftchainservice

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math/big"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"go.uber.org/zap"

	"github.com/ProjectsTask/EasySwapBase/logger/xzap"
)

const fetchIPFSTimeout = 30 * time.Second

var (
	hosts = []string{"https://ipfs.io/ipfs/", "https://cf-ipfs.com/ipfs/", "https://infura-ipfs.io/ipfs/", "https://cloudflare-ipfs.com/ipfs/"}
)

type nftInfoSimple struct {
	Name         string `json:"name"`
	Image        string `json:"image"`
	Description  string `json:"description"`
	ExternalLink string `json:"external_link"`
	Attributes   interface{}
}

// fetchNftMetadata 获取NFT元数据
// 该方法根据collectionAddr和tokenID参数，从区块链或IPFS网络中获取NFT的元数据
// 它首先通过调用智能合约的方法获取tokenURI，然后根据URI的前缀判断元数据的存储位置和格式
// 如果是base64编码的JSON数据，它将解码数据；如果是IPFS上的数据，它将调用fetchIpfsData方法获取数据
// 最后，它将获取到的元数据返回给调用者
func (s *Service) fetchNftMetadata(collectionAddr string, tokenID string) ([]byte, string, error) {
	beginTime := time.Now()
	// 记录获取NFT元数据开始的日志
	xzap.WithContext(s.ctx).Info("fetch nft metadata start",
		zap.String("collection_addr", collectionAddr), zap.String("token_id", tokenID), zap.Time("start", beginTime))
	defer func() {
		// 记录获取NFT元数据结束的日志，并计算耗时
		xzap.WithContext(s.ctx).Info("fetch nft metadata end", zap.String("collection_addr", collectionAddr), zap.String("token_id", tokenID), zap.Float64("take", time.Now().Sub(beginTime).Seconds()))
	}()
	// 将tokenID转换为大整数类型
	tokenId, _ := big.NewInt(0).SetString(tokenID, 10)
	// 调用智能合约的tokenURI方法获取tokenURI的数据
	tokenURIReqData, err := s.Abi.Pack("tokenURI", tokenId)
	if err != nil {
		return nil, "", errors.Wrap(err, fmt.Sprintf("failed on pack token uri %s", tokenID))
	}

	to := common.HexToAddress(collectionAddr)
	// 通过以太坊节点调用智能合约，获取tokenURI的响应数据
	respData, err := s.NodeClient.CallContract(s.ctx, ethereum.CallMsg{To: &to, Data: tokenURIReqData}, nil)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed on request token uri")
	}
	// 解析tokenURI的响应数据
	res, err := s.Abi.Unpack("tokenURI", respData)
	if err != nil {
		return nil, "", errors.Wrap(err, "failed on unpack token uri")
	}

	tokenUri := res[0].(string)
	var body []byte
	// 根据tokenUri的前缀判断元数据的存储位置和格式，并获取元数据
	if len(tokenUri) > 29 && tokenUri[0:29] == "data:application/json;base64," {
		// 解码base64编码的JSON数据
		body, err = base64.StdEncoding.DecodeString(tokenUri[29:])
		if err != nil {
			return nil, "", errors.Wrap(err, fmt.Sprintf("failed on decode token uri: %s", tokenUri))
		}
	} else if len(tokenUri) > 5 && tokenUri[0:5] == "ipfs:" {
		// 从IPFS网络获取数据
		body, err = s.fetchIpfsData(tokenUri)
		if err != nil {
			return nil, "", errors.Wrap(err, fmt.Sprintf("failed on fetch token uri: %s", tokenUri))
		}
	} else if len(tokenUri) > 5 && tokenUri[0:4] != "http" {
		// 如果URI格式不正确，返回错误
		return nil, "", errors.New(fmt.Sprintf("invalid url %s", tokenUri))
	}
	// 如果是HTTP链接，获取JSON格式的元数据
	if len(tokenUri) > 5 && tokenUri[0:4] == "http" {
		body, err = s.fetchJsonData(tokenUri)
		if err != nil {
			return nil, "", errors.Wrap(err, fmt.Sprintf("failed on fetch metadata. uri:%s", tokenUri))
		}
	}
	// 处理获取到的元数据，去除可能的BOM标记
	if body != nil {
		body = bytes.TrimPrefix(body, []byte("\xef\xbb\xbf"))
		return body, tokenUri, nil
	}
	// 如果元数据为空，返回错误
	return nil, "", errors.New("empty metadata")
}

// FetchNftOwner 获取NFT所有者地址。
// 该函数根据收藏品地址和代币ID来查询并返回NFT的所有者地址。
// 参数:
//
//	collectionAddr - NFT收藏品的合约地址。
//	tokenID - NFT代币的唯一标识符。
//
// 返回值:
//
//	common.Address - NFT所有者的以太坊地址。
//	error - 在查询过程中可能发生的错误。
func (s *Service) FetchNftOwner(collectionAddr string, tokenID string) (common.Address, error) {
	// 记录开始时间以监控函数执行效率。
	beginTime := time.Now()
	// 日志记录：函数开始执行。
	xzap.WithContext(s.ctx).Info("fetch nft owner start",
		zap.String("collection_addr", collectionAddr), zap.String("token_id", tokenID), zap.Time("start", beginTime))
	// 延迟执行，记录函数结束和执行时间。
	defer func() {
		xzap.WithContext(s.ctx).Info("fetch nft owner end", zap.String("collection_addr", collectionAddr), zap.String("token_id", tokenID), zap.Float64("take", time.Now().Sub(beginTime).Seconds()))
	}()
	// 将代币ID转换为大整数类型，以便在以太坊网络上进行操作
	tokenId, _ := big.NewInt(0).SetString(tokenID, 10)
	// 构建查询所有者的请求数据。
	tokenOwnerReqData, err := s.Abi.Pack("ownerOf", tokenId)
	if err != nil {
		// 错误处理：在构建请求数据时出错。
		return common.Address{}, errors.Wrap(err, fmt.Sprintf("failed on pack token uri %s", tokenID))
	}
	// 将收藏品地址转换为以太坊地址类型。
	to := common.HexToAddress(collectionAddr)
	// 调用以太坊节点客户端查询合约，获取所有者信息。
	respData, err := s.NodeClient.CallContract(s.ctx, ethereum.CallMsg{To: &to, Data: tokenOwnerReqData}, nil)
	if err != nil {
		// 错误处理：在请求所有者信息时出错。
		return common.Address{}, errors.Wrap(err, "failed on request token uri")
	}
	// 解析响应数据，提取所有者地址。
	res, err := s.Abi.Unpack("ownerOf", respData)
	if err != nil {
		// 错误处理：在解析响应数据时出错。
		return common.Address{}, errors.Wrap(err, "failed on unpack token uri")
	}
	// 将解析出的所有者地址转换为以太坊地址类型并返回。
	address := *abi.ConvertType(res[0], new(common.Address)).(*common.Address)
	return address, nil
}

// fetchIpfsData 通过IPFS获取数据。
// 该方法尝试从多个IPFS主机获取数据，直到成功或超时。
// 参数:tokenUri - 包含IPFS路径的URI。
// 返回值：
//
//	[]byte - 获取到的数据。
//	error - 如果获取失败，返回错误信息。
func (s *Service) fetchIpfsData(tokenUri string) ([]byte, error) {
	// 创建一个通道，用于接收成功获取的数据。
	finished := make(chan []byte)
	// 创建一个定时器，用于在fetchIPFSTimeout后触发超时。
	ticker := time.NewTicker(fetchIPFSTimeout)
	// 定义一个取消函数切片，用于取消未完成的HTTP请求。
	var cancelFns []context.CancelFunc
	// 遍历所有IPFS主机，尝试获取数据
	for i := range hosts {
		host := hosts[i]
		// 构建完整的IPFS请求URL。
		fullUrl := strings.Replace(tokenUri, "ipfs://", host, 1)
		// 创建一个可取消的上下文，用于控制HTTP请求的生命周期。
		ctx, cancel := context.WithCancel(context.Background())
		// 将取消函数添加到切片中，以便后续取消所有未完成的请求。
		cancelFns = append(cancelFns, cancel)
		// 启动一个新的goroutine，用于发起HTTP请求。
		go func() {
			// 创建HTTP GET请求。
			req, err := http.NewRequestWithContext(ctx, "GET", fullUrl, nil)
			if err != nil {
				// 记录错误日志，并退出goroutine。
				xzap.WithContext(s.ctx).Error("failed on create fetch ipfs data req",
					zap.String("token_uri", tokenUri), zap.Error(err))
				return
			}
			// 发送HTTP请求，并获取响应。
			resp, err := s.HttpClient.Do(req)
			if err != nil {
				// 记录错误日志，并退出goroutine。
				xzap.WithContext(s.ctx).Error("failed on fetch metadata from ipfs",
					zap.String("token_uri", tokenUri), zap.Error(err))
				return
			}
			// 检查HTTP响应状态码是否为200（成功）。
			if resp.StatusCode == http.StatusOK {
				// 读取HTTP响应体。
				body, err := io.ReadAll(resp.Body)
				if err != nil {
					// 记录错误日志，并退出goroutine。
					xzap.WithContext(s.ctx).Error("failed on read ipfs req resp body",
						zap.String("token_uri", tokenUri), zap.Error(err))
					return
				}
				// 将获取到的数据发送到finished通道。
				finished <- body
			}
		}()
	}
	// 监听finished通道和定时器，等待数据获取成功或超时。
	select {
	case data := <-finished:
		// 数据获取成功，取消所有未完成的请求，并返回获取到的数据。
		for i := range cancelFns {
			cancelFns[i]()
		}

		return data, nil
	case <-ticker.C:
		// 超时，返回错误信息。
		return nil, errors.New("request metadata timeout: " + tokenUri)
	}
}

func (s *Service) fetchJsonData(tokenUri string) (body []byte, err error) {
	resp, err := s.HttpClient.Get(tokenUri)
	if err != nil {
		return nil, errors.Wrap(err, "failed on get metadata from http")
	}

	if resp.StatusCode == http.StatusOK {
		body, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed on read resp body")
		}

		if strings.Contains(tokenUri, "squid-app-o5c27.ondigitalocean") {
			tmpData := struct {
				Msg  string        `json:"msg"`
				Data nftInfoSimple `json:"data"`
			}{}
			if err := json.Unmarshal(body, &tmpData); err != nil {
				return nil, errors.Wrap(err, "failed on unmarshal raw metadata")
			}

			if tmpData.Data.Name != "" {
				body, err = json.Marshal(tmpData.Data)
				if err != nil {
					return nil, errors.Wrap(err, "failed on marshal raw metadata")
				}
			}
		}
	}

	return body, err
}

func (s *Service) FetchOnChainMetadata(collectionAddr string, tokenID string) (*JsonMetadata, error) {
	rawData, tokenUri, err := s.fetchNftMetadata(collectionAddr, tokenID)
	if err != nil {
		return nil, errors.Wrap(err, "failed on fetch nft metadata")
	}

	if len(rawData) == 0 {
		return nil, errors.New("metadata length is zero")
	}

	metadata, err := DecodeJsonMetadata(rawData, tokenUri, s.NameTags, s.ImageTags, s.AttributesTags, s.TraitNameTags, s.TraitValueTags)
	if err != nil {
		return nil, errors.Wrap(err, "failed on decode metadata")
	}

	return metadata, nil
}

// DecodeJsonMetadata parses the NFT token Metadata JSON.
func DecodeJsonMetadata(content []byte, tokenUri string, nameTags, imageTags, attributesTags, traitNameTags, traitValueTags []string) (*JsonMetadata, error) {
	_, err := url.Parse(string(content))
	if err == nil {
		return &JsonMetadata{
			Image: string(content),
		}, nil
	}

	if isImageFile(content) {
		return &JsonMetadata{
			Image: tokenUri,
		}, nil
	}

	if isTextFile(content) {
		metadatas := make(map[string]interface{})
		err := json.Unmarshal(content, &metadatas)
		if err != nil {
			return nil, errors.Wrap(err, "failed on unmarshal raw metadata")
		}

		var metadata JsonMetadata
		// parse name field
		for _, tag := range nameTags {
			v, ok := metadatas[tag]
			if ok {
				sname, ok := v.(string)
				if ok {
					metadata.Name = sname
				}

				fname, ok := v.(float64)
				if ok {
					metadata.Name = strconv.FormatFloat(fname, 'f', -1, 64)
				}

				if metadata.Name != "" {
					break
				}
			}
		}
		// parse image field
		for _, tag := range imageTags {
			v, ok := metadatas[tag]
			if ok {
				simage, ok := v.(string)
				if ok {
					metadata.Image = simage
				}
				if metadata.Image != "" {
					break
				}
			}
		}
		// parse attributes field
		for _, tag := range attributesTags {
			rawAttributes, ok := metadatas[tag]
			if ok {
				attributesArray, ok := rawAttributes.(map[string]interface{})
				if ok {
					for k, v := range attributesArray {
						var value string
						svalue, ok := v.(string)
						if ok {
							value = svalue
						}
						fvalue, ok := v.(float64)
						if ok {
							value = strconv.FormatFloat(fvalue, 'f', -1, 64)
						}

						metadata.Attributes = append(metadata.Attributes, &OpenseaMetadataProps{
							TraitType: k,
							Value:     value,
						})
					}
					break
				}

				attributesArrays, ok := rawAttributes.([]interface{})
				if ok {
					for _, attributes := range attributesArrays {
						attributesMap, ok := attributes.(map[string]interface{})
						if !ok {
							break
						}
						var trait string
						for _, tag := range traitNameTags {
							v, ok := attributesMap[tag]
							if ok {
								strait, ok := v.(string)
								if ok {
									trait = strait
								}
								ftrait, ok := v.(float64)
								if ok {
									trait = strconv.FormatFloat(ftrait, 'f', -1, 64)
								}

								break
							}
						}
						var value string
						for _, tag := range traitValueTags {
							v, ok := attributesMap[tag]
							if ok {
								svalue, ok := v.(string)
								if ok {
									value = svalue
								}
								fvalue, ok := v.(float64)
								if ok {
									value = strconv.FormatFloat(fvalue, 'f', -1, 64)
								}

								break
							}
						}
						if trait != "" && value != "" {
							metadata.Attributes = append(metadata.Attributes, &OpenseaMetadataProps{
								TraitType: trait,
								Value:     value,
							})
						}
					}
					break
				}
			}
		}
		return &metadata, nil
	}

	return nil, errors.New(fmt.Sprintf("unsupported content type:%s", string(content)))
}

// isTextFile returns true if file content format is plain text or empty.
func isTextFile(data []byte) bool {
	if len(data) == 0 {
		return true
	}
	return strings.Contains(http.DetectContentType(data), "text/")
}

func isImageFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "image/")
}

func isVideoFile(data []byte) bool {
	return strings.Contains(http.DetectContentType(data), "video/")
}
