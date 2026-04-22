// SPDX-License-Identifier: BUSL-1.1
/**
 * @title  IAnchoredCompliance
 * @notice Comprehensive interface for the AnchoredCompliance contract containing all functions, events, and errors
 */
pragma solidity ^0.8.20;

import { IAccessControl } from "@openzeppelin/contracts/access/IAccessControl.sol";
import { IAnchoredBlocklist } from "contracts/interfaces/IAnchoredBlocklist.sol";
import { IAnchoredSanctionsList } from "contracts/interfaces/IAnchoredSanctionsList.sol";

interface IAnchoredCompliance is IAccessControl {
    // ============ Functions ============

    /**
     * @notice Check if a user is compliant with the Anchored token's compliance rules
     * @param  anchoredToken The Anchored token address
     * @param  user     The user address
     * @dev    Reverts if the user is not compliant. Does not revert for unsupported tokens.
     */
    function checkIsCompliant(address anchoredToken, address user) external view;

    /**
     * @notice Simplified compliance check using _msgSender() as Anchored identifier
     * @param user The user address to check
     */
    function checkIsCompliant(address user) external view;

    /**
     * @notice Set the blocklist for a given Anchored token
     * @param  anchoredToken  The Anchored token address
     * @param  blocklist The blocklist contract
     */
    function setBlocklist(address anchoredToken, IAnchoredBlocklist blocklist) external;

    /**
     * @notice Set the sanctions list for a given Anchored token
     * @param  anchoredToken      The Anchored token address
     * @param  sanctionsList The sanctions list contract
     */
    function setSanctionsList(address anchoredToken, IAnchoredSanctionsList sanctionsList) external;

    /**
     * @notice Set the role for an Anchored token
     * @param  anchoredToken The Anchored token address
     * @dev    This role is computed as the keccak256 hash of the Anchored token address.
     */
    function setAnchoredRole(address anchoredToken) external;

    // ============ View Functions ============

    // Note: hasRole, getRoleAdmin, getRoleMember, getRoleMemberCount, and DEFAULT_ADMIN_ROLE are inherited from IAccessControl

    /**
     * @notice Returns the MASTER_CONFIGURER_ROLE constant
     * @return The bytes32 value of MASTER_CONFIGURER_ROLE
     */
    function MASTER_CONFIGURE_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the ANCHORED_ROLE_MANAGER constant
     * @return The bytes32 value of ANCHORED_ROLE_MANAGER
     */
    function ANCHORED_MANAGE_ROLE() external view returns (bytes32);

    /**
     * @notice Returns the role for a specific Anchored token
     * @param anchoredToken The Anchored token address
     * @return The role bytes32 for the token
     */
    function anchoredRole(address anchoredToken) external view returns (bytes32);

    /**
     * @notice Returns the blocklist contract for a specific Anchored token
     * @param anchoredToken The Anchored token address
     * @return The blocklist contract
     */
    function anchoredTokenToBlocklist(address anchoredToken) external view returns (IAnchoredBlocklist);

    /**
     * @notice Returns the sanctions list contract for a specific Anchored token
     * @param anchoredToken The Anchored token address
     * @return The sanctions list contract
     */
    function anchoredTokenToSanctionsList(address anchoredToken) external view returns (IAnchoredSanctionsList);

    // ============ Events ============

    /**
     * @notice Emitted when the role is set for an Anchored token
     * @param  anchoredToken The Anchored token address
     * @param  role     The role that was set - keccak256 hash of the Anchored token address
     */
    event AnchoredRoleSet(address indexed anchoredToken, bytes32 role);

    /**
     * @notice Emitted when the blocklist is set for an Anchored token
     * @param  anchoredToken     The Anchored token address for which the blocklist was set
     * @param  oldBlocklist The old blocklist contract
     * @param  newBlocklist The new blocklist contract
     */
    event BlocklistSet(address indexed anchoredToken, IAnchoredBlocklist oldBlocklist, IAnchoredBlocklist newBlocklist);

    /**
     * @notice Emitted when the sanctions list is set for an Anchored token
     * @param  anchoredToken         The Anchored token address for which the sanctions list was set
     * @param  oldSanctionsList The old sanctions list contract
     * @param  newSanctionsList The new sanctions list contract
     */
    event SanctionsListSet(
        address indexed anchoredToken, IAnchoredSanctionsList oldSanctionsList, IAnchoredSanctionsList newSanctionsList
    );

    // ============ Errors ============

    /// Error thrown when a user is blocked via the blocklist
    error UserBlocked();

    /// Error thrown when a user is sanctioned via the sanctions list
    error UserSanctioned();

    /// Error thrown when trying to set an Anchored token address of 0x0
    error AnchoredAddressCannotBeZero();

    /// Error thrown when trying to set configure an Anchored token without the required role
    error MissingAnchoredOrMasterConfigurerRole();
}
