// SPDX-License-Identifier: MIT
pragma solidity ^0.8.7;

contract SimpleStorage {
    uint256 private value;
    
    event ValueChanged(uint256 newValue, address changedBy);
    
    function setValue(uint256 _value) public {
        value = _value;
        emit ValueChanged(_value, msg.sender);
    }
    
    function getValue() public view returns (uint256) {
        return value;
    }
}
