// SPDX-License-Identifier: MIT
pragma solidity ^0.8.19;

import "hardhat/console.sol";

contract DebugContract {
    uint256 public value;
    
    constructor() {
        console.log("DebugContract deployed by:", msg.sender);
        console.log("Block number:", block.number);
    }
    
    function setValue(uint256 newValue) public {
        console.log("Setting value from %s to %s", value, newValue);
        console.log("Called by:", msg.sender);
        
        value = newValue;
        
        console.log("Value set to:", value);
    }
    
    function complexFunction(uint256 a, uint256 b) public pure returns (uint256) {
        console.log("Inputs: a=%s, b=%s", a, b);
        
        uint256 result = a + b;
        console.log("Addition result:", result);
        
        result = result * 2;
        console.log("Multiplied result:", result);
        
        if (result > 100) {
            console.log("Result is greater than 100");
        } else {
            console.log("Result is 100 or less");
        }
        
        return result;
    }
}
