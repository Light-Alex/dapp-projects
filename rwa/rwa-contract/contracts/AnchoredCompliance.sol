// SPDX-License-Identifier: BUSL-1.1
pragma solidity ^0.8.20;

import { IAnchoredCompliance } from "contracts/interfaces/IAnchoredCompliance.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { IAnchoredBlocklist } from "contracts/interfaces/IAnchoredBlocklist.sol";
import { IAnchoredSanctionsList } from "contracts/interfaces/IAnchoredSanctionsList.sol";

/**
 * @title  AnchoredCompliance
 * @notice This contract is responsible for enforcing compliance rules for Anchored tokens and
 *         associated systems. It allows for setting blocklists and
 *         sanctions lists for Anchored tokens.
 *         Roles:
 *          - MASTER_CONFIGURE_ROLE
 *          - ANCHORED_MANAGE_ROLE
 */
contract AnchoredCompliance is IAnchoredCompliance, AccessControlEnumerable, Initializable {
    /// Role to set the blocklist or sanctions list for any Anchored token
    bytes32 public constant MASTER_CONFIGURE_ROLE = keccak256("MASTER_CONFIGURE_ROLE");

    /// Role admin for roles for setting blocklists and sanctions lists for specific Anchored tokens
    bytes32 public constant ANCHORED_MANAGE_ROLE = keccak256("ANCHORED_MANAGE_ROLE");

    /**
     * @notice Mapping of Anchored token address to the role that can set the blocklist or sanctions list
     *         for specific Anchored tokens
     */
    mapping(address /*anchoredToken*/ => bytes32) public anchoredRole;

    /// Mapping of Anchored token address to the blocklist contract
    mapping(address => IAnchoredBlocklist) public anchoredTokenToBlocklist;

    /// Mapping of Anchored token address to the sanctions list contract
    mapping(address => IAnchoredSanctionsList) public anchoredTokenToSanctionsList;

    /**
     * @notice Constructor for implementation contract
     */
    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param admin_ The address that will be granted the default admin role
     */
    function initialize(address admin_) external initializer {
        _grantRole(DEFAULT_ADMIN_ROLE, admin_);
    }

    /**
     * @notice Check if a user is compliant with the Anchored token's compliance rules
     * @param  anchoredToken The Anchored token address
     * @param  user     The user address
     * @dev    Reverts if the user is not compliant. Does not revert for unsupported tokens.
     */
    function checkIsCompliant(address anchoredToken, address user) external view override {
        if (
            address(anchoredTokenToBlocklist[anchoredToken]) != address(0)
                && anchoredTokenToBlocklist[anchoredToken].isBlocked(user)
        ) {
            revert UserBlocked();
        }

        if (
            address(anchoredTokenToSanctionsList[anchoredToken]) != address(0)
                && anchoredTokenToSanctionsList[anchoredToken].isSanctioned(user)
        ) revert UserSanctioned();
    }

    /**
     * @notice Set the blocklist for a given Anchored token
     * @param  anchoredToken  The Anchored token address
     * @param  blocklist The blocklist contract
     */
    function setBlocklist(address anchoredToken, IAnchoredBlocklist blocklist) external {
        if (!(hasRole(anchoredRole[anchoredToken], _msgSender()) || hasRole(MASTER_CONFIGURE_ROLE, _msgSender()))) {
            revert MissingAnchoredOrMasterConfigurerRole();
        }
        if (anchoredToken == address(0)) revert AnchoredAddressCannotBeZero();
        emit BlocklistSet(anchoredToken, anchoredTokenToBlocklist[anchoredToken], blocklist);

        anchoredTokenToBlocklist[anchoredToken] = blocklist;
    }

    /**
     * @notice Set the sanctions list for a given Anchored token
     * @param  anchoredToken      The Anchored token address
     * @param  sanctionsList The sanctions list contract
     */
    function setSanctionsList(address anchoredToken, IAnchoredSanctionsList sanctionsList) external {
        if (!(hasRole(anchoredRole[anchoredToken], _msgSender()) || hasRole(MASTER_CONFIGURE_ROLE, _msgSender()))) {
            revert MissingAnchoredOrMasterConfigurerRole();
        }
        if (anchoredToken == address(0)) revert AnchoredAddressCannotBeZero();
        emit SanctionsListSet(anchoredToken, anchoredTokenToSanctionsList[anchoredToken], sanctionsList);

        anchoredTokenToSanctionsList[anchoredToken] = sanctionsList;
    }

    /**
     * @notice Set the role for an Anchored token
     * @param  anchoredToken The Anchored token address
     * @dev    This role is computed as the keccak256 hash of the Anchored token address.
     */
    function setAnchoredRole(address anchoredToken) external onlyRole(ANCHORED_MANAGE_ROLE) {
        if (anchoredToken == address(0)) revert AnchoredAddressCannotBeZero();
        bytes32 role = keccak256(abi.encodePacked(anchoredToken));
        _setRoleAdmin(role, ANCHORED_MANAGE_ROLE);
        anchoredRole[anchoredToken] = role;

        emit AnchoredRoleSet(anchoredToken, role);
    }

    /**
     * @notice Simplified compliance check using _msgSender() as Anchored identifier
     * @param user The user address to check
     */
    function checkIsCompliant(address user) external view {
        this.checkIsCompliant(_msgSender(), user);
    }
}
