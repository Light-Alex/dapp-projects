// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {ContextUpgradeable} from "@openzeppelin/contracts-upgradeable/utils/ContextUpgradeable.sol";
import {EIP712Upgradeable} from "@openzeppelin/contracts-upgradeable/utils/cryptography/EIP712Upgradeable.sol";
//价格类型，通常用于订单的价格字段
import {Price} from "./libraries/RedBlackTreeLibrary.sol";
//LibOrder定义订单相关的结构和工具函数,OrderKey订单的唯一标识符
import {LibOrder, OrderKey} from "./libraries/LibOrder.sol";

/**
 * @title Verify the validity of the order parameters.
 */
// 抽象合约，用于验证订单的有效性、管理订单的状态（如已成交数量或取消状态），
// 并提供与订单相关的辅助功能
abstract contract OrderValidator is
    Initializable, //支持合约的初始化功能，通常用于可升级合约
    ContextUpgradeable, //提供上下文信息
    EIP712Upgradeable //实现 EIP-712 标准，用于结构化数据的签名和验证。
{
    bytes4 private constant EIP_1271_MAGIC_VALUE = 0x1626ba7e;
    //表示订单已取消的特殊值，等于 uint256 的最大值
    uint256 private constant CANCELLED = type(uint256).max;

    // fillsStat record orders filled status, key is the order hash,
    // and value is filled amount.
    // Value CANCELLED means the order has been canceled.
    // 记录订单的已成交数量
    mapping(OrderKey => uint256) public filledAmount;

    //初始化合约，设置 EIP-712 的名称和版本.部署合约时调用
    function __OrderValidator_init(
        string memory EIP712Name,
        string memory EIP712Version
    ) internal onlyInitializing {
        __Context_init();
        __EIP712_init(EIP712Name, EIP712Version);
        __OrderValidator_init_unchained();
    }

    function __OrderValidator_init_unchained() internal onlyInitializing {}

    /**
     * @notice Validate order parameters.
     * @param order  Order to validate.
     * @param isSkipExpiry  Skip expiry check if true.
     */
    //验证订单的基本参数是否有效
    function _validateOrder(
        LibOrder.Order memory order,
        bool isSkipExpiry
    ) internal view {
        // Order must have a maker.必须有合法的 maker 地址
        require(order.maker != address(0), "OVa: miss maker");
        // Order must be started and not be expired.

        if (!isSkipExpiry) {
            // Skip expiry check if true.如果未跳过过期检查，确保订单未过期
            require(
                order.expiry == 0 || order.expiry > block.timestamp,
                "OVa: expired"
            );
        }
        // Order salt cannot be 0.值必须非零
        require(order.salt != 0, "OVa: zero salt");

        if (order.side == LibOrder.Side.List) {
            // 卖单（List），必须指定合法的 NFT 集合地址
            require(
                order.nft.collection != address(0),
                "OVa: unsupported nft asset"
            );
        } else if (order.side == LibOrder.Side.Bid) {
            //买单（Bid），价格必须大于零
            require(Price.unwrap(order.price) > 0, "OVa: zero price");
        }
    }

    /**
     * @notice Get filled amount of orders.
     * @param orderKey  The hash of the order.
     * @return orderFilledAmount Has completed fill amount of sell order (0 if order is unfilled).
     */
    // 获取订单已成交数量
    function _getFilledAmount(
        OrderKey orderKey
    ) internal view returns (uint256 orderFilledAmount) {
        // Get has completed fill amount.
        orderFilledAmount = filledAmount[orderKey];
        // Cancelled order cannot be matched.
        //如果订单已取消，抛出异常
        require(orderFilledAmount != CANCELLED, "OVa: canceled");
    }

    /**
     * @notice Update filled amount of orders.
     * @param newAmount  New fill amount of order.
     * @param orderKey  The hash of the order.
     */
    //更新订单的已成交数量
    function _updateFilledAmount(
        uint256 newAmount,
        OrderKey orderKey
    ) internal {
        //如果新数量等于 CANCELLED，抛出异常
        require(newAmount != CANCELLED, "OVa: canceled");
        filledAmount[orderKey] = newAmount;
    }

    /**
     * @notice Cancel order.
     * @dev Cancelled orders cannot be reopened.
     * @param orderKey  The hash of the order.
     */
    //将订单标记为已取消
    function _cancelOrder(OrderKey orderKey) internal {
        //已取消的订单无法重新打开或匹配
        filledAmount[orderKey] = CANCELLED;
    }

    uint256[50] private __gap;
}
