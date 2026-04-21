// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";
// 提供安全的 ETH 和 NFT 转移方法,ERC-721 标准接口
import {LibTransferSafeUpgradeable, IERC721} from "./libraries/LibTransferSafeUpgradeable.sol";
// 定义订单相关的结构和工具函数
import {LibOrder, OrderKey} from "./libraries/LibOrder.sol";

import {IEasySwapVault} from "./interface/IEasySwapVault.sol";

// EasySwapVault 是一个功能全面的保管库合约，支持 ETH 和 NFT 的存储、提取、编辑和转移，
// 主要用于支持 EasySwapOrderBook 的订单匹配和资产管理功能
contract EasySwapVault is IEasySwapVault, OwnableUpgradeable {
    // IEasySwapVault：实现保管库的接口。
    // OwnableUpgradeable：提供所有权管理功能，支持升级的合约。
    using LibTransferSafeUpgradeable for address;
    using LibTransferSafeUpgradeable for IERC721;
    //存储 EasySwapOrderBook 合约的地址，只有该地址可以调用保管库的核心功能
    address public orderBook;
    // 记录每个订单的 ETH 余额
    mapping(OrderKey => uint256) public ETHBalance;
    // 记录每个订单的 NFT 信息（存储 tokenId）
    mapping(OrderKey => uint256) public NFTBalance;
    // 确保只有 EasySwapOrderBook 合约可以调用
    modifier onlyEasySwapOrderBook() {
        require(msg.sender == orderBook, "HV: only EasySwap OrderBook");
        _;
    }

    // 初始化合约，设置合约所有者
    function initialize() public initializer {
        __Ownable_init(_msgSender());
    }

    // 设置或更新 EasySwapOrderBook 合约地址,只有合约所有者可以调用
    function setOrderBook(address newOrderBook) public onlyOwner {
        require(newOrderBook != address(0), "HV: zero address");
        orderBook = newOrderBook;
    }

    // 查询指定订单的 ETH 和 NFT 余额
    function balanceOf(
        OrderKey orderKey
    ) external view returns (uint256 ETHAmount, uint256 tokenId) {
        ETHAmount = ETHBalance[orderKey];
        tokenId = NFTBalance[orderKey];
    }

    // 接收 ETH 并存储到指定订单的余额中
    function depositETH(
        OrderKey orderKey,
        uint256 ETHAmount
    ) external payable onlyEasySwapOrderBook {
        require(msg.value >= ETHAmount, "HV: not match ETHAmount");
        ETHBalance[orderKey] += msg.value;
    }

    // 从指定订单的余额中提取 ETH，并转移到指定地址
    function withdrawETH(
        OrderKey orderKey,
        uint256 ETHAmount,
        address to
    ) external onlyEasySwapOrderBook {
        ETHBalance[orderKey] -= ETHAmount;
        to.safeTransferETH(ETHAmount);
    }

    //接收 NFT 并存储到指定订单的余额中
    function depositNFT(
        OrderKey orderKey,
        address from,
        address collection,
        uint256 tokenId
    ) external onlyEasySwapOrderBook {
        IERC721(collection).safeTransferNFT(from, address(this), tokenId);

        NFTBalance[orderKey] = tokenId;
    }

    //从指定订单的余额中提取 NFT，并转移到指定地址
    function withdrawNFT(
        OrderKey orderKey,
        address to,
        address collection,
        uint256 tokenId
    ) external onlyEasySwapOrderBook {
        // 验证 tokenId 是否匹配
        require(NFTBalance[orderKey] == tokenId, "HV: not match tokenId");
        delete NFTBalance[orderKey];

        IERC721(collection).safeTransferNFT(address(this), to, tokenId);
    }

    // 编辑订单的 ETH 余额
    function editETH(
        OrderKey oldOrderKey,
        OrderKey newOrderKey,
        uint256 oldETHAmount,
        uint256 newETHAmount,
        address to
    ) external payable onlyEasySwapOrderBook {
        ETHBalance[oldOrderKey] = 0;
        if (oldETHAmount > newETHAmount) {
            //如果新订单的 ETH 金额小于旧订单，退还多余的 ETH
            ETHBalance[newOrderKey] = newETHAmount;
            to.safeTransferETH(oldETHAmount - newETHAmount);
        } else if (oldETHAmount < newETHAmount) {
            //如果新订单的 ETH 金额大于旧订单，接收额外的 ETH
            require(
                msg.value >= newETHAmount - oldETHAmount,
                "HV: not match newETHAmount"
            );
            ETHBalance[newOrderKey] = msg.value + oldETHAmount;
        } else {
            //如果金额相等，直接迁移余额
            ETHBalance[newOrderKey] = oldETHAmount;
        }
    }

    //将 NFT 从旧订单迁移到新订单
    function editNFT(
        OrderKey oldOrderKey,
        OrderKey newOrderKey
    ) external onlyEasySwapOrderBook {
        NFTBalance[newOrderKey] = NFTBalance[oldOrderKey];
        delete NFTBalance[oldOrderKey];
    }

    //转移指定的 ERC-721 资产
    function transferERC721(
        address from,
        address to,
        LibOrder.Asset calldata assets
    ) external onlyEasySwapOrderBook {
        IERC721(assets.collection).safeTransferNFT(from, to, assets.tokenId);
    }

    //批量转移多个 ERC-721 资产
    function batchTransferERC721(
        address to,
        LibOrder.NFTInfo[] calldata assets
    ) external {
        for (uint256 i = 0; i < assets.length; ++i) {
            IERC721(assets[i].collection).safeTransferNFT(
                _msgSender(),
                to,
                assets[i].tokenId
            );
        }
    }

    //实现 ERC-721 接口的 onERC721Received 方法，用于接收 NFT
    function onERC721Received(
        address,
        address,
        uint256,
        bytes memory
    ) public virtual returns (bytes4) {
        return this.onERC721Received.selector;
    }

    //接收 ETH
    receive() external payable {}

    uint256[50] private __gap;
}
