// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

import { IAnchoredTokenLike } from "contracts/interfaces/IAnchoredLike.sol";
import { IAnchoredTokenManager } from "contracts/interfaces/IAnchoredTokenManager.sol";
import { AccessControlEnumerable } from "@openzeppelin/contracts/access/extensions/AccessControlEnumerable.sol";
import { Initializable } from "@openzeppelin/contracts/proxy/utils/Initializable.sol";
import { ReentrancyGuard } from "@openzeppelin/contracts/utils/ReentrancyGuard.sol";
import { IERC20 } from "@openzeppelin/contracts/token/ERC20/IERC20.sol";
import { SafeERC20 } from "@openzeppelin/contracts/token/ERC20/utils/SafeERC20.sol";
import { ECDSA } from "@openzeppelin/contracts/utils/cryptography/ECDSA.sol";
import { MessageHashUtils } from "@openzeppelin/contracts/utils/cryptography/MessageHashUtils.sol";

/**
 * @title  AnchoredTokenManager
 * @notice This contract manages the minting and redemption of Anchored tokens
 */
contract AnchoredTokenManager is IAnchoredTokenManager, AccessControlEnumerable, ReentrancyGuard, Initializable {
    using SafeERC20 for IERC20;
    using ECDSA for bytes32;

    /// @notice Role identifier for those who can configure the contract
    bytes32 public constant CONFIGURE_ROLE = keccak256("CONFIGURE_ROLE");

    /// @notice Role identifier for those who can pause minting
    bytes32 public constant PAUSE_ROLE = keccak256("PAUSE_ROLE");

    /// @notice Role identifier for those who can sign quotes
    bytes32 public constant QUOTE_SIGN_ROLE = keccak256("QUOTE_SIGN_ROLE");

    /// @notice Role identifier for those who can update multipliers
    bytes32 public constant MULTIPLIER_UPDATE_ROLE = keccak256("MULTIPLIER_UPDATE_ROLE");

    /// @notice Normalizer for Anchored token quantities
    uint256 public constant ANCHORED_TOKEN_NORMALIZER = 1e18;

    /// @notice The Anchored identifier used for compliance checks
    bytes32 public immutable ANCHORED_STOCK_TOKEN_IDENTIFIER;

    /// @notice The address of the USDC token
    address public immutable USDC;

    /**
     * @notice Minimum Usd amount required to subscribe to an Anchored token, denoted in Usd with 18
     *         decimals of precision
     */
    uint256 public minimumDepositUsd;

    /**
     * @notice Minimum Usd amount required to redeem an Anchored token, denoted in Usd with 18
     *         decimals of precision
     */
    uint256 public minimumRedemptionUsd;

    /// @notice Mapping to track executed attestation IDs to prevent replay attacks
    mapping(bytes32 => bool) public executedAttestationIds;

    /// @notice Mapping to track whether minting is paused for a specific Anchored token
    mapping(address => bool) public anchoredTokenMintingPaused;

    /// @notice Mapping to track whether redemptions are paused for a specific Anchored token
    mapping(address => bool) public anchoredTokenRedemptionsPaused;

    /// @notice Map of all Anchored tokens that are accepted for minting and redemptions
    mapping(address => bool) public anchoredTokenAccepted;

    /**
     * @notice Constructor for implementation contract
     * @param _usdc The address of the USDC token
     * @param _anchoredStockTokenIdentifier The Anchored identifier
     */
    constructor(address _usdc, bytes32 _anchoredStockTokenIdentifier) {
        _disableInitializers();

        // Initialize immutable variables
        USDC = _usdc;
        ANCHORED_STOCK_TOKEN_IDENTIFIER = _anchoredStockTokenIdentifier;
    }

    /**
     * @notice Initialize function for proxy deployment
     * @param usdc_ The address of the USDC token
     * @param guardian_ The address which will be granted admin and other roles
     * @param minimumDepositUsd_ Minimum Usd amount required to subscribe
     * @param minimumRedemptionUsd_ Minimum Usd amount required to redeem
     */
    function initialize(address usdc_, address guardian_, uint256 minimumDepositUsd_, uint256 minimumRedemptionUsd_)
        external
        initializer
    {
        if (usdc_ == address(0)) revert UsdcAddressCannotBeZero();
        if (guardian_ == address(0)) revert GuardianAddressCannotBeZero();

        minimumDepositUsd = minimumDepositUsd_;
        minimumRedemptionUsd = minimumRedemptionUsd_;

        _grantRole(DEFAULT_ADMIN_ROLE, guardian_);
        _grantRole(CONFIGURE_ROLE, guardian_);
        _grantRole(PAUSE_ROLE, guardian_);
        _grantRole(QUOTE_SIGN_ROLE, guardian_);
        _grantRole(MULTIPLIER_UPDATE_ROLE, guardian_);
    }

    /**
     * @notice Modifier to check if minting is not paused for a token
     * @param  token The token address to check
     */
    modifier whenMintingNotPaused(address token) {
        if (anchoredTokenMintingPaused[token]) revert MintingPaused();
        _;
    }

    /**
     * @notice Modifier to check if redemption is not paused for a token
     * @param  token The token address to check
     */
    modifier whenRedeemNotPaused(address token) {
        if (anchoredTokenRedemptionsPaused[token]) revert RedemptionPaused();
        _;
    }

    /**
     * @notice Called by users to mint Anchored tokens with USDC
     * @param  quote                The quote to mint Anchored tokens with
     * @param  signature            The signature of the quote attestation
     * @param  depositToken         The token the user is depositing (must be USDC)
     * @param  depositTokenAmount   The amount of deposit tokens (must equal USDC value)
     * @return The amount of Anchored tokens minted
     */
    function mintWithAttestation(
        Quote calldata quote,
        bytes memory signature,
        address depositToken,
        uint256 depositTokenAmount
    ) public nonReentrant whenMintingNotPaused(quote.asset) returns (uint256) {
        if (depositToken != address(USDC)) revert InvalidDepositToken();
        if (quote.side != QuoteSide.BUY) revert InvalidQuoteSide();
        if (!anchoredTokenAccepted[quote.asset]) revert TokenNotAccepted();

        _verifyQuote(quote, signature);
        executedAttestationIds[quote.attestationId] = true;

        // Usd values are normalized to 18 decimals
        uint256 mintUsdValue = (quote.quantity * quote.price) / ANCHORED_TOKEN_NORMALIZER;
        if (mintUsdValue < minimumDepositUsd) revert DepositAmountTooSmall();
        if (depositTokenAmount != mintUsdValue) revert InvalidDepositAmount();

        // Transfer USDC into the contract (no burn needed for USDC)
        IERC20(USDC).safeTransferFrom(_msgSender(), address(this), mintUsdValue);

        // Mint Anchored tokens to the user
        IAnchoredTokenLike(quote.asset).mint(_msgSender(), quote.quantity);

        _emitTradeExecuted(quote);

        return quote.quantity;
    }

    /**
     * @notice Called by users to redeem Anchored tokens for USDC
     * @param  quote                The quote to redeem Anchored tokens with
     * @param  signature            The signature of the quote attestation
     * @param  receiveToken         The token the user would like to receive (must be USDC)
     * @return The amount of USDC transferred
     */
    function redeemWithAttestation(Quote calldata quote, bytes memory signature, address receiveToken)
        public
        nonReentrant
        whenRedeemNotPaused(quote.asset)
        returns (uint256)
    {
        if (receiveToken != address(USDC)) revert InvalidReceiveToken();
        if (quote.side != QuoteSide.SELL) revert InvalidQuoteSide();
        if (!anchoredTokenAccepted[quote.asset]) revert TokenNotAccepted();

        _verifyQuote(quote, signature);
        executedAttestationIds[quote.attestationId] = true;

        // Burn the Anchored tokens
        IERC20(quote.asset).safeTransferFrom(_msgSender(), address(this), quote.quantity);
        IAnchoredTokenLike(quote.asset).burn(quote.quantity);

        // Usd values are normalized to 18 decimals
        uint256 redemptionUsdcValue = (quote.quantity * quote.price) / ANCHORED_TOKEN_NORMALIZER;

        if (redemptionUsdcValue < minimumRedemptionUsd) {
            revert RedemptionAmountTooSmall();
        }

        // Transfer USDC to the user
        IERC20(USDC).safeTransfer(_msgSender(), redemptionUsdcValue);

        _emitTradeExecuted(quote);

        return redemptionUsdcValue;
    }

    /**
     * @notice Verifies the signature of the quote
     * @param  quote     The quote to verify
     * @param  signature The signature to verify
     */
    function _verifyQuote(Quote calldata quote, bytes memory signature) internal view {
        if (executedAttestationIds[quote.attestationId]) {
            revert AttestationAlreadyExecuted();
        }
        if (quote.expiry < block.timestamp) revert QuoteExpired();

        bytes32 messageHash = keccak256(
            abi.encode(
                quote.asset, quote.side, quote.quantity, quote.price, quote.nonce, quote.expiry, quote.attestationId
            )
        );

        address signer = ECDSA.recover(MessageHashUtils.toEthSignedMessageHash(messageHash), signature);
        if (!hasRole(QUOTE_SIGN_ROLE, signer)) revert InvalidSigner();
    }

    /**
     * @notice Emits a trade executed event
     * @param  quote The quote that was executed
     */
    function _emitTradeExecuted(Quote calldata quote) internal {
        emit TradeExecuted(quote.asset, _msgSender(), quote.side, quote.quantity, quote.price);
    }

    /**
     * @notice Sets whether a token is accepted for mints and redemptions on this contract
     * @param  token    The address of the token
     * @param  accepted Whether the token is accepted for mints and redemptions
     */
    function setAnchoredTokenRegistrationStatus(address token, bool accepted) external onlyRole(CONFIGURE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        anchoredTokenAccepted[token] = accepted;
        emit AnchoredTokenRegistered(token, accepted);
    }

    /**
     * @notice Sets the minimum amount required for a subscription
     * @param  minimumDepositUsd_ The minimum amount required to subscribe
     */
    function setMinimumDepositAmount(uint256 minimumDepositUsd_) external onlyRole(CONFIGURE_ROLE) {
        emit MinimumDepositAmountSet(minimumDepositUsd, minimumDepositUsd_);
        minimumDepositUsd = minimumDepositUsd_;
    }

    /**
     * @notice Sets the minimum amount to redeem
     * @param  minimumRedemptionUsd_ The minimum amount to redeem
     */
    function setMinimumRedemptionAmount(uint256 minimumRedemptionUsd_) external onlyRole(CONFIGURE_ROLE) {
        emit MinimumRedemptionAmountSet(minimumRedemptionUsd, minimumRedemptionUsd_);
        minimumRedemptionUsd = minimumRedemptionUsd_;
    }

    /**
     * @notice Pauses minting for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function pauseAnchoredTokenMinting(address token) external onlyRole(PAUSE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        anchoredTokenMintingPaused[token] = true;
        emit AnchoredTokenMintingPaused(token);
    }

    /**
     * @notice Unpauses minting for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function unpauseAnchoredTokenMinting(address token) external onlyRole(PAUSE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        anchoredTokenMintingPaused[token] = false;
        emit AnchoredTokenMintingUnpaused(token);
    }

    /**
     * @notice Pauses redemptions for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function pauseAnchoredTokenRedemptions(address token) external onlyRole(PAUSE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        anchoredTokenRedemptionsPaused[token] = true;
        emit AnchoredTokenRedemptionsPaused(token);
    }

    /**
     * @notice Unpauses redemptions for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function unpauseAnchoredTokenRedemptions(address token) external onlyRole(PAUSE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        anchoredTokenRedemptionsPaused[token] = false;
        emit AnchoredTokenRedemptionsUnpaused(token);
    }

    /**
     * @notice Updates the multiplier for a specific Anchored token
     * @param  token The address of the Anchored token
     * @param  newMultiplier The new multiplier value
     */
    function updateMultiplier(address token, uint256 newMultiplier) external onlyRole(MULTIPLIER_UPDATE_ROLE) {
        if (token == address(0)) revert TokenAddressCannotBeZero();
        if (!anchoredTokenAccepted[token]) revert TokenNotAccepted();

        // Call the updateMultiplier function on the Anchored token
        IAnchoredTokenLike(token).updateMultiplier(newMultiplier);

        emit MultiplierUpdated(token, newMultiplier);
    }
}
