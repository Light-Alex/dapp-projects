pragma solidity ^0.6.0;

import "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import "@openzeppelin/contracts/math/SafeMath.sol";
import "@openzeppelin/contracts/cryptography/ECDSA.sol";


contract Airdrop {

    using ECDSA for bytes32;
    using SafeMath for *;

    IERC20 public airdropToken;
    uint256 public totalTokensWithdrawn;

    mapping (address => bool) public wasClaimed;
    uint256 public constant TOKENS_PER_CLAIM = 100 * 10**18;

    event TokensAirdropped(address beneficiary);

    // Constructor, initial setup
    constructor(address _airdropToken) public {
        require(_airdropToken != address(0));

        airdropToken = IERC20(_airdropToken);
    }

    // Function to withdraw tokens.
    function withdrawTokens() public {
        // 确保调用者是交易原始发起者（始终是外部账户 EOA），而不是合约
        require(msg.sender == tx.origin, "Require that message sender is tx-origin.");

        address beneficiary = msg.sender;

        require(!wasClaimed[beneficiary], "Already claimed!");
        wasClaimed[msg.sender] = true;

        bool status = airdropToken.transfer(beneficiary, TOKENS_PER_CLAIM);
        require(status, "Token transfer status is false.");

        totalTokensWithdrawn = totalTokensWithdrawn.add(TOKENS_PER_CLAIM);
        emit TokensAirdropped(beneficiary);
    }

}
