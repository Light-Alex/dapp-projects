// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

// 导入 Chainlink 的 IDecimalAggregator 接口
import "@chainlink/contracts/src/v0.8/data-feeds/interfaces/IDecimalAggregator.sol";
import "@chainlink/contracts/src/v0.8/data-feeds/interfaces/ICommonAggregator.sol";
// @chainlink\contracts\src\v0.8\data-feeds\interfaces\IDecimalAggregator.sol

/**
 * @title EthPriceFeed
 * @dev 一个演示如何使用 Chainlink Data Feeds 获取 ETH/USD 价格的合约
 * @dev 主要用于教学演示，生产环境需考虑更多安全性和错误处理
 */
contract EthPriceFeed {
    // 声明一个 IDecimalAggregator 类型的状态变量，用于存储价格喂价合约的地址
    IDecimalAggregator internal priceFeed;

    // 在特定区块链上，ETH/USD 价格喂价合约的地址是预定义的，可以在 Chainlink 官方文档中找到
    // 不同网络的 Chainlink 价格喂价合约地址不同，以下是一些示例：
    // Sepolia 测试网: 0x694AA1769357215DE4FAC081bf1f309aDC325306
    // Goerli 测试网: 0xD4a33860578De61DBAbDc8BFdb98FD742fA7028e
    // Ethereum 主网: 0x5f4eC3Df9cbd43714FE2740f5E3616155c5b8419

    /**
     * @dev 构造函数，初始化价格喂价合约地址
     * @param _priceFeedAddress 目标区块链网络上 Chainlink ETH/USD 价格喂价合约的地址
     */
    constructor(address _priceFeedAddress) {
        priceFeed = IDecimalAggregator(_priceFeedAddress);
    }

    /**
     * @dev 获取最新的 ETH/USD 价格
     * @return price 返回最新价格 (int256)
     * @notice 价格的小数位数通常为 8，可通过 getDecimals() 函数查询
     */
    function getLatestPrice() public view returns (int256) {
        // 调用 IDecimalAggregator 的 latestRoundData 函数获取最新一轮的价格数据
        // 该函数返回多个值，此处我们只关心 `price`
        (
            uint80 roundId, 
            int256 answer, 
            uint256 startedAt, 
            uint256 updatedAt, 
            uint80 answeredInRound
        ) = priceFeed.latestRoundData();
        
        // 返回 ETH/USD 价格
        return answer;
    }

    /**
     * @dev 获取价格数据的小数位数
     * @return 返回价格值的小数位数 (uint8)
     * @notice 了解小数位数对于正确解析价格值至关重要
     */
    function getDecimals() public view returns (uint8) {
        return priceFeed.decimals();
    }

    /**
     * @dev 获取价格喂价的描述信息
     * @return 返回描述该价格喂对的字符串 (例如 "ETH / USD")
     */
    function getDescription() public view returns (string memory) {
        return ICommonAggregator(address(priceFeed)).description();
    }

    /**
     * @dev 一个辅助函数，将获取的价格格式化为更易读的数值
     * @return 返回经过格式化处理的价格 (uint256)
     * @notice 此函数将价格除以 10^decimals 来得到带有所需小数位的数值
     */
    function getFormattedPrice() public view returns (uint256) {
        int256 price = getLatestPrice();
        uint8 decimals = getDecimals();
        // 将价格调整到正确的小数点位置
        return uint256(price) / (10 ** uint256(decimals));
    }
}
