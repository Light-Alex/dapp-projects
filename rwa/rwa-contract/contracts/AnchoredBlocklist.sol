/**
 * SPDX-License-Identifier: BUSL-1.1
 */
pragma solidity ^0.8.20;

import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { IAnchoredBlocklist } from "contracts/interfaces/IAnchoredBlocklist.sol";

/**
 * @title Blocklist
 * @notice This contract manages the blocklist status for accounts using granular role-based access control.
 */
contract AnchoredBlocklist is AccessControlEnumerable, IAnchoredBlocklist, Initializable {
    // Role constants
    bytes32 public constant BLOCKLIST_ADD_ROLE = keccak256("BLOCKLIST_ADD_ROLE");
    bytes32 public constant BLOCKLIST_REMOVE_ROLE = keccak256("BLOCKLIST_REMOVE_ROLE");

    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param admin_ The address which will be granted admin and other roles
     */
    function initialize(address admin_) external initializer {
        _grantRole(DEFAULT_ADMIN_ROLE, admin_);
        _grantRole(BLOCKLIST_ADD_ROLE, admin_);
        _grantRole(BLOCKLIST_REMOVE_ROLE, admin_);
    }

    // {<address> => is account blocked}
    mapping(address => bool) private blockedAddresses;

    /**
     * @notice Function to add a list of accounts to the blocklist
     * @dev Can be called by DEFAULT_ADMIN_ROLE or BLOCKLIST_ADD_ROLE
     * @param accounts Array of addresses to block
     */
    function addToBlocklist(address[] calldata accounts) external onlyRole(BLOCKLIST_ADD_ROLE) {
        for (uint256 i; i < accounts.length; ++i) {
            if (accounts[i] == address(0)) revert BlocklistAddAddressCannotBeZero();
            blockedAddresses[accounts[i]] = true;
        }
        emit BlockedAddressesAdded(accounts);
    }

    /**
     * @notice Function to remove a list of accounts from the blocklist
     * @dev Can be called by DEFAULT_ADMIN_ROLE or BLOCKLIST_REMOVE_ROLE
     * @param accounts Array of addresses to unblock
     */
    function removeFromBlocklist(address[] calldata accounts) external onlyRole(BLOCKLIST_REMOVE_ROLE) {
        for (uint256 i; i < accounts.length; ++i) {
            if (accounts[i] == address(0)) revert BlocklistRemoveAddressCannotBeZero();
            blockedAddresses[accounts[i]] = false;
        }
        emit BlockedAddressesRemoved(accounts);
    }

    /**
     * @notice Function to check if an account is blocked
     *
     * @param addr Address to check
     *
     * @return True if account is blocked, false otherwise
     */
    function isBlocked(address addr) external view returns (bool) {
        return blockedAddresses[addr];
    }
}
