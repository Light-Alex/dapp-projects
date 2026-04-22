// SPDX-License-Identifier: BUSL-1.1
/**
 * @title  IAnchoredSanctionsList
 * @notice Comprehensive interface for the AnchoredSanctionsList contract containing all functions, events, and errors
 */
pragma solidity ^0.8.20;

import { IAccessControl } from "@openzeppelin/contracts/access/IAccessControl.sol";

interface IAnchoredSanctionsList is IAccessControl {
    // ============ Functions ============

    /**
     * @notice Add addresses to the sanctions list
     * @param newSanctions Array of addresses to add to sanctions list
     */
    function addToSanctionsList(address[] memory newSanctions) external;

    /**
     * @notice Remove addresses from the sanctions list
     * @param removeSanctions Array of addresses to remove from sanctions list
     */
    function removeFromSanctionsList(address[] memory removeSanctions) external;

    // ============ View Functions ============

    /**
     * @notice Check if an address is sanctioned
     * @param addr Address to check
     * @return True if the address is sanctioned
     */
    function isSanctioned(address addr) external view returns (bool);

    /**
     * @notice Check if an address is sanctioned with verbose output
     * @param addr Address to check
     * @return True if the address is sanctioned
     */
    function isSanctionedVerbose(address addr) external returns (bool);

    // ============ Role Constants ============

    /**
     * @notice Returns the SANCTIONS_ADD_ROLE constant
     * @return The bytes32 value of SANCTIONS_ADD_ROLE
     */
    function SANCTIONS_ADD_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the SANCTIONS_REMOVE_ROLE constant
     * @return The bytes32 value of SANCTIONS_REMOVE_ROLE
     */
    function SANCTIONS_REMOVE_ROLE() external view returns (bytes32);

    // ============ Events ============

    /**
     * @notice Event emitted when an address is sanctioned
     * @param addr The sanctioned address
     */
    event SanctionedAddress(address indexed addr);

    /**
     * @notice Event emitted when an address is not sanctioned
     * @param addr The non-sanctioned address
     */
    event NonSanctionedAddress(address indexed addr);

    /**
     * @notice Event emitted when addresses are added to the sanctions list
     * @param addrs The addresses that were added to the sanctions list
     */
    event SanctionedAddressesAdded(address[] addrs);

    /**
     * @notice Event emitted when addresses are removed from the sanctions list
     * @param addrs The addresses that were removed from the sanctions list
     */
    event SanctionedAddressesRemoved(address[] addrs);

    // ============ Errors ============

    /// Error thrown when attempting to add zero address to sanctions list
    error SanctionsListAddAddressCannotBeZero();

    /// Error thrown when attempting to remove zero address from sanctions list
    error SanctionsListRemoveAddressCannotBeZero();
}
