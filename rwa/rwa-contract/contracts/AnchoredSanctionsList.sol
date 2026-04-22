/**
 * SPDX-License-Identifier: BUSL-1.1
 */
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { IAnchoredSanctionsList } from "contracts/interfaces/IAnchoredSanctionsList.sol";

pragma solidity ^0.8.20;

contract AnchoredSanctionsList is IAnchoredSanctionsList, AccessControlEnumerable, Initializable {
    // Role constants
    bytes32 public constant SANCTIONS_ADD_ROLE = keccak256("SANCTIONS_ADD_ROLE");
    bytes32 public constant SANCTIONS_REMOVE_ROLE = keccak256("SANCTIONS_REMOVE_ROLE");

    mapping(address => bool) private sanctionedAddresses;

    constructor() {
        _disableInitializers();
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param admin_ The address which will be granted admin and other roles
     */
    function initialize(address admin_) external initializer {
        _grantRole(DEFAULT_ADMIN_ROLE, admin_);
        _grantRole(SANCTIONS_ADD_ROLE, admin_);
        _grantRole(SANCTIONS_REMOVE_ROLE, admin_);
    }

    function addToSanctionsList(address[] memory newSanctions) public onlyRole(SANCTIONS_ADD_ROLE) {
        for (uint256 i = 0; i < newSanctions.length; i++) {
            if (newSanctions[i] == address(0)) revert SanctionsListAddAddressCannotBeZero();
            sanctionedAddresses[newSanctions[i]] = true;
        }
        emit SanctionedAddressesAdded(newSanctions);
    }

    function removeFromSanctionsList(address[] memory removeSanctions) public onlyRole(SANCTIONS_REMOVE_ROLE) {
        for (uint256 i = 0; i < removeSanctions.length; i++) {
            if (removeSanctions[i] == address(0)) revert SanctionsListRemoveAddressCannotBeZero();
            sanctionedAddresses[removeSanctions[i]] = false;
        }
        emit SanctionedAddressesRemoved(removeSanctions);
    }

    function isSanctioned(address addr) public view returns (bool) {
        return sanctionedAddresses[addr] == true;
    }

    function isSanctionedVerbose(address addr) public returns (bool) {
        if (isSanctioned(addr)) {
            emit SanctionedAddress(addr);
            return true;
        } else {
            emit NonSanctionedAddress(addr);
            return false;
        }
    }
}
