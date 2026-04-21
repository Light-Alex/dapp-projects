// SPDX-License-Identifier: MIT

pragma solidity ^0.8.19;

import {Initializable} from "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";
import {OwnableUpgradeable} from "@openzeppelin/contracts-upgradeable/access/OwnableUpgradeable.sol";

import {LibPayInfo} from "./libraries/LibPayInfo.sol";

/// @title 协议管理合约
/// @notice 管理协议分成比例的核心合约，支持可升级初始化
abstract contract ProtocolManager is
    Initializable, // 可初始化合约
    OwnableUpgradeable // 可升级的权限控制
{
    uint128 public protocolShare; // 协议分成比例
    // 当协议分成比例更新时触发
    event LogUpdatedProtocolShare(uint128 indexed newProtocolShare);

    function __ProtocolManager_init(
        uint128 newProtocolShare
    ) internal onlyInitializing {
        // __Ownable_init(_msgSender());
        __ProtocolManager_init_unchained(newProtocolShare);
    }

    function __ProtocolManager_init_unchained(
        uint128 newProtocolShare
    ) internal onlyInitializing {
        _setProtocolShare(newProtocolShare);
    }

    // 设置协议分成比例（外部调用）
    function setProtocolShare(uint128 newProtocolShare) external onlyOwner {
        _setProtocolShare(newProtocolShare);
    }

    // 设置协议分成比例
    function _setProtocolShare(uint128 newProtocolShare) internal {
        require(
            newProtocolShare <= LibPayInfo.MAX_PROTOCOL_SHARE,
            "PM: exceed max protocol share"
        );
        protocolShare = newProtocolShare;
        emit LogUpdatedProtocolShare(newProtocolShare);
    }

    uint256[50] private __gap; //保留升级空间
}
