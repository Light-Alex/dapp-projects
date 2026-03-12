// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "./MyUpgradeable.sol";

contract MyUpgradeableV2 is MyUpgradeableV1 {
    // 新增功能
    function multiply(uint256 factor) public {
        setValue(getValue() * factor);
    }
}