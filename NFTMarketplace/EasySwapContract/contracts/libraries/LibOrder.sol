// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import {Price} from "./RedBlackTreeLibrary.sol";

type OrderKey is bytes32;

library LibOrder {
    enum Side {
        List,
        Bid
    }

    enum SaleKind {
        FixedPriceForCollection,
        FixedPriceForItem
    }

    struct Asset {
        uint256 tokenId;
        address collection;
        uint96 amount;
    }

    struct NFTInfo {
        address collection;
        uint256 tokenId;
    }

    struct Order {
        Side side; // 订单类型（List 或 Bid）
        SaleKind saleKind; // 销售类型（固定价格或其他）
        address maker; // 创建者地址
        Asset nft; // NFT 信息
        Price price; // unit price of nft 单个 NFT 的价格
        uint64 expiry; // 过期时间
        uint64 salt; // 随机数
    }

    struct DBOrder {
        Order order;
        OrderKey next;
    }

    /// @dev Order queue: used to store orders of the same price
    struct OrderQueue {
        OrderKey head; // 队列头部
        OrderKey tail; // 队列尾部
    }

    struct EditDetail {
        OrderKey oldOrderKey; // old order key which need to be edit
        LibOrder.Order newOrder; // new order struct which need to be add
    }

    struct MatchDetail {
        LibOrder.Order sellOrder;
        LibOrder.Order buyOrder;
    }

    OrderKey public constant ORDERKEY_SENTINEL = OrderKey.wrap(0x0);

    bytes32 public constant ASSET_TYPEHASH =
        keccak256("Asset(uint256 tokenId,address collection,uint96 amount)");

    bytes32 public constant ORDER_TYPEHASH =
        keccak256(
            "Order(uint8 side,uint8 saleKind,address maker,Asset nft,uint128 price,uint64 expiry,uint64 salt)Asset(uint256 tokenId,address collection,uint96 amount)"
        );

    // 用于生成订单的唯一标识符
    function hash(Asset memory asset) internal pure returns (bytes32) {
        return
            keccak256(
                abi.encode(
                    ASSET_TYPEHASH,
                    asset.tokenId,
                    asset.collection,
                    asset.amount
                )
            );
    }

    function hash(Order memory order) internal pure returns (OrderKey) {
        return
            OrderKey.wrap(
                keccak256(
                    abi.encodePacked(
                        ORDER_TYPEHASH,
                        order.side,
                        order.saleKind,
                        order.maker,
                        hash(order.nft),
                        Price.unwrap(order.price),
                        order.expiry,
                        order.salt
                    )
                )
            );
    }

    // 检查是否到达队列头尾，检查给定的 OrderKey 是否等于预定义的哨兵节点 ORDERKEY_SENTINEL。
    function isSentinel(OrderKey orderKey) internal pure returns (bool) {
        return OrderKey.unwrap(orderKey) == OrderKey.unwrap(ORDERKEY_SENTINEL);
    }

    // 检查是否到达队列头尾
    function isNotSentinel(OrderKey orderKey) internal pure returns (bool) {
        return OrderKey.unwrap(orderKey) != OrderKey.unwrap(ORDERKEY_SENTINEL);
    }
}
