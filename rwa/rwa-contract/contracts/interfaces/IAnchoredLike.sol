/**
 * SPDX-License-Identifier: BUSL-1.1
 */
pragma solidity ^0.8.20;

// This interface is not inherited directly by Anchored, instead, it is a
// subset of functions provided by all Anchored tokens that the Anchored Hub
// Client uses.
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";

interface IAnchoredTokenLike is IERC20 {
    function mint(address to, uint256 amount) external;

    function burn(uint256 amount) external;

    function burnFrom(address from, uint256 amount) external;

    function updateMultiplier(uint256 newMultiplier) external;
}
