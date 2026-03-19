// SPDX-License-Identifier: MIT
pragma solidity ^0.8.20;

import "@openzeppelin/contracts/token/ERC20/ERC20.sol";
import "@openzeppelin/contracts/utils/Strings.sol";
import "@openzeppelin/contracts/utils/Address.sol";
import "@openzeppelin/contracts/access/Ownable.sol";
import "@openzeppelin/contracts/utils/ReentrancyGuard.sol";

contract Airdrop is Ownable, ReentrancyGuard {

    IERC20 public airdropToken; 
    mapping(address=>bool) public isGov; 
    uint batchSize = 500;
    uint256 airdropAmount = 0;
    uint processNum = 0;
    uint8 public decimals = 18;

    event Airdropped(
        address indexed addr,
        uint256 amount,
        uint256 processIndex,
        uint256 timestamp,
        string symbol
    );
   
    constructor(IERC20 _mtkToken) Ownable(msg.sender) { 
        airdropToken = _mtkToken;
    }

    modifier onlyOwnerOrGov(){
        require(isGov[msg.sender] || msg.sender == owner(), "not the gov or owner");
        _;
    }

    function setGov(address addr)external onlyOwner {
        isGov[addr] = true;
    }

    function removeGov(address addr)external onlyOwner {
        isGov[addr] = false;
    }

    function airdropERC20(address[] memory recipients, uint256[] memory amounts) external onlyOwnerOrGov {
        require(recipients.length == amounts.length, "length of address doesn't match with length of amounts");
        string memory symbol;
        try IERC20Metadata(address(airdropToken)).symbol() returns (string memory _symbol) {
            symbol = _symbol;
        } catch {
            symbol = "MTK";
        }

        for(uint i = 0; i < recipients.length; i++){
            address addr = recipients[i];
            uint256 amount = amounts[i];
            require(amount > 0, "Invalid amount");
            require(addr != address(0), "Invalid address");
            require(airdropToken.transferFrom(msg.sender, addr, amount),"transfer failed") ;
            emit Airdropped(addr, amount, i, block.timestamp, symbol);
        }
   }

    function airdropBNB(address[] memory recipients, uint256[] memory amounts) external payable onlyOwnerOrGov nonReentrant{
        require(recipients.length == amounts.length, "length of address doesn't match with length of amounts");
        uint256 totalAmount = getSumAmount(amounts);
        require(msg.value == totalAmount, string.concat("balance of BNB is not enough, msg.vaule=", Strings.toString(msg.value) , ", totalAmount=", Strings.toString(totalAmount)));

        for(uint i = 0; i < recipients.length; i++){
            address payable addr = payable(recipients[i]);
            uint256 amount=amounts[i];

            require(amount > 0, "Invalid amount");
            require(addr != address(0), "Invalid address");

            (bool success, ) = addr.call{value: amount}("");
            require(success,"transfer failed");

            emit Airdropped(addr, amount, i, block.timestamp, "BNB");
        }       
    }


    /**
     * @dev 接收BNB（必须实现，否则合约无法接收BNB）
     */
    receive() external payable {}

    function getSumAmount( uint256[] memory amounts) internal pure returns (uint256){
        uint256 totalAmount = 0;
          for(uint256 i = 0; i<amounts.length; i++){
            totalAmount += amounts[i];
        }
        return totalAmount;
    }
   
}