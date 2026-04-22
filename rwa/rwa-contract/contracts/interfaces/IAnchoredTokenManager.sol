// SPDX-License-Identifier: UNLICENSED
/**
 * Copyright (c) 2023, Anchored Finance
 */
pragma solidity ^0.8.20;

/**
 * @title  IAnchoredTokenManager
 * @notice Comprehensive interface for the AnchoredTokenManager contract containing all functions, events, and errors
 */
interface IAnchoredTokenManager {
    // ============ Enums ============

    enum QuoteSide {
        /// Indicates that the user is buying Anchored tokens
        BUY,
        /// Indicates that the user is selling Anchored tokens
        SELL
    }

    // ============ Structs ============

    /**
     * @notice Quote struct that is signed by the attestation signer
     * @param  attestationId  The ID of the quote
     * @param  asset          The address of the Anchored token being bought or sold
     * @param  price          The price of the Anchored token in Usd with 18 decimals
     * @param  quantity       The quantity of Anchored tokens being bought or sold
     * @param  expiry         The expiration of the quote in seconds since the epoch
     * @param  side           The direction of the quote (BUY or SELL)
     * @param  nonce          The nonce for the quote
     */
    struct Quote {
        bytes32 attestationId;
        address asset;
        uint256 price;
        uint256 quantity;
        uint256 expiry;
        QuoteSide side;
        uint256 nonce;
    }

    // ============ Functions ============

    /**
     * @notice Called by users to mint Anchored tokens with USDC
     * @param  quote                The quote to mint Anchored tokens with
     * @param  signature            The signature of the quote attestation
     * @param  depositToken         The token the user is depositing (must be USDC)
     * @param  depositTokenAmount   The amount of deposit tokens (must equal USDC value)
     * @return receivedAnchoredTokenAmount The amount of Anchored tokens minted
     */
    function mintWithAttestation(
        Quote calldata quote,
        bytes memory signature,
        address depositToken,
        uint256 depositTokenAmount
    ) external returns (uint256 receivedAnchoredTokenAmount);

    /**
     * @notice Called by users to redeem Anchored tokens for USDC
     * @param  quote                The quote to redeem Anchored tokens with
     * @param  signature            The signature of the quote attestation
     * @param  receiveToken         The token the user would like to receive (must be USDC)
     * @return The amount of USDC transferred
     */
    function redeemWithAttestation(Quote calldata quote, bytes memory signature, address receiveToken)
        external
        returns (uint256);

    /**
     * @notice Sets whether a token is accepted for mints and redemptions on this contract
     * @param  token    The address of the token
     * @param  accepted Whether the token is accepted for mints and redemptions
     */
    function setAnchoredTokenRegistrationStatus(address token, bool accepted) external;

    /**
     * @notice Sets the minimum amount required for a subscription
     * @param  minimumDepositUsd_ The minimum amount required to subscribe
     */
    function setMinimumDepositAmount(uint256 minimumDepositUsd_) external;

    /**
     * @notice Sets the minimum amount to redeem
     * @param  minimumRedemptionUsd_ The minimum amount to redeem
     */
    function setMinimumRedemptionAmount(uint256 minimumRedemptionUsd_) external;

    /**
     * @notice Pauses minting for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function pauseAnchoredTokenMinting(address token) external;

    /**
     * @notice Unpauses minting for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function unpauseAnchoredTokenMinting(address token) external;

    /**
     * @notice Pauses redemptions for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function pauseAnchoredTokenRedemptions(address token) external;

    /**
     * @notice Unpauses redemptions for a specific Anchored token
     * @param  token The address of the Anchored token
     */
    function unpauseAnchoredTokenRedemptions(address token) external;

    /**
     * @notice Updates the multiplier for a specific Anchored token
     * @param  token The address of the Anchored token
     * @param  newMultiplier The new multiplier value
     */
    function updateMultiplier(address token, uint256 newMultiplier) external;

    /**
     * @notice Initialize function for proxy deployment
     * @param usdc_ The address of the USDC token
     * @param guardian_ The address which will be granted admin and other roles
     * @param minimumDepositUsd_ Minimum Usd amount required to subscribe
     * @param minimumRedemptionUsd_ Minimum Usd amount required to redeem
     */
    function initialize(address usdc_, address guardian_, uint256 minimumDepositUsd_, uint256 minimumRedemptionUsd_)
        external;

    // ============ View Functions ============

    /**
     * @notice Returns whether an attestation ID has been executed
     * @param attestationId The attestation ID to check
     * @return True if the attestation has been executed
     */
    function executedAttestationIds(bytes32 attestationId) external view returns (bool);

    /**
     * @notice Returns whether minting is paused for a specific Anchored token
     * @param token The address of the Anchored token
     * @return True if minting is paused for the token
     */
    function anchoredTokenMintingPaused(address token) external view returns (bool);

    /**
     * @notice Returns whether redemptions are paused for a specific Anchored token
     * @param token The address of the Anchored token
     * @return True if redemptions are paused for the token
     */
    function anchoredTokenRedemptionsPaused(address token) external view returns (bool);

    /**
     * @notice Returns whether an Anchored token is accepted for minting and redemptions
     * @param token The address of the Anchored token
     * @return True if the token is accepted
     */
    function anchoredTokenAccepted(address token) external view returns (bool);

    // ============ Events ============

    /**
     * @notice Event emitted when a trade is executed with an attestation
     * @param  asset          The address of the Anchored token being bought or sold
     * @param  user           The user executing the trade
     * @param  side           The direction of the quote (BUY or SELL)
     * @param  quantity       The quantity of Anchored tokens being bought or sold
     * @param  price          The price of the Anchored token in Usd with 18 decimals
     */
    event TradeExecuted(address asset, address user, QuoteSide side, uint256 quantity, uint256 price);

    /**
     * @notice Event emitted when subscription minimum is set
     * @param  oldMinDepositAmount Old subscription minimum
     * @param  newMinDepositAmount New subscription minimum
     */
    event MinimumDepositAmountSet(uint256 indexed oldMinDepositAmount, uint256 indexed newMinDepositAmount);

    /**
     * @notice Event emitted when redeem minimum is set
     * @param  oldMinRedemptionAmount Old redeem minimum
     * @param  newMinRedemptionAmount New redeem minimum
     */
    event MinimumRedemptionAmountSet(uint256 indexed oldMinRedemptionAmount, uint256 indexed newMinRedemptionAmount);

    /**
     * @notice Event emitted when the accepted Anchored token is set
     * @param  anchoredToken    The address of the Anchored token
     * @param  registered Whether the Anchored token is registered
     */
    event AnchoredTokenRegistered(address indexed anchoredToken, bool indexed registered);

    /**
     * @notice Event emitted when minting is paused for a specific Anchored token
     * @param anchoredToken The address of the Anchored token
     */
    event AnchoredTokenMintingPaused(address indexed anchoredToken);

    /**
     * @notice Event emitted when minting is unpaused for a specific Anchored token
     * @param anchoredToken The address of the Anchored token
     */
    event AnchoredTokenMintingUnpaused(address indexed anchoredToken);

    /**
     * @notice Event emitted when redemption is paused for a specific Anchored token
     * @param anchoredToken The address of the Anchored token
     */
    event AnchoredTokenRedemptionsPaused(address indexed anchoredToken);

    /**
     * @notice Event emitted when redemption is unpaused for a specific Anchored token
     * @param anchoredToken The address of the Anchored token
     */
    event AnchoredTokenRedemptionsUnpaused(address indexed anchoredToken);

    /**
     * @notice Event emitted when multiplier is updated for a specific Anchored token
     * @param token The address of the Anchored token
     * @param newMultiplier The new multiplier value
     */
    event MultiplierUpdated(address indexed token, uint256 indexed newMultiplier);

    // ============ Errors ============

    /// Error emitted when the token address is zero
    error TokenAddressCannotBeZero();

    /// Error emitted when the deposit amount is too small
    error DepositAmountTooSmall();

    /// @notice Thrown when the attestation has already been executed
    error AttestationAlreadyExecuted();

    /// @notice Thrown when the quote has expired
    error QuoteExpired();

    /// @notice Thrown when the signer is invalid
    error InvalidSigner();

    /// @notice Thrown when the token is not accepted
    error TokenNotAccepted();

    /// @notice Thrown when minting is paused
    error MintingPaused();

    /// @notice Thrown when redemption is paused
    error RedemptionPaused();

    /// @notice Thrown when the deposit token is invalid
    error InvalidDepositToken();

    /// @notice Thrown when the receive token is invalid
    error InvalidReceiveToken();

    /// @notice Thrown when the deposit amount is invalid
    error InvalidDepositAmount();

    /// @notice Thrown when the guardian address is zero
    error GuardianAddressCannotBeZero();

    /// Error emitted when the redemption amount is too small
    error RedemptionAmountTooSmall();

    /// Custom error for invalid quote direction
    error InvalidQuoteSide();

    /// Error emitted when the Anchored Token is not registered for minting/redemption
    error AnchoredTokenNotRegistered();

    /// Error emitted when attempting to set the `USDC` address to zero
    error UsdcAddressCannotBeZero();
}
