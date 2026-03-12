// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts-upgradeable/proxy/utils/Initializable.sol";

contract MyUpgradeableV1 is Initializable {
    uint256 private _value;
    
    // 使用initialize代替constructor
    function initialize(uint256 initialValue) public initializer {
        _value = initialValue;
    }
    
    function getValue() public view returns (uint256) {
        return _value;
    }
    
    function setValue(uint256 newValue) public {
        _value = newValue;
    }
}
