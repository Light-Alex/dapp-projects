# RWA (Real World Asset) System Architecture (Consolidated)

> 📖 **Related Documentation**: For detailed repository design and service directory structure, please refer to [REPO_DESIGN.md](./docs/REPO_DESIGN.md)

## Overview

This document describes the RWA (Real World Asset) system architecture based on the Alpaca trading API. The system enables tokenization of traditional financial assets (e.g., equities) and supports core functions such as deposits, event listening, and order placement.

## System Roles

Based on the flow analysis, the system includes the following core roles:

1. **Trader** - End user who initiates trading requests
2. **RWA Market** - Core marketplace that handles tokenized asset trading
3. **MM (Market Maker)** - Provides liquidity
4. **Anchored** - Asset anchoring and token issuance party
5. **Broker** - Traditional brokerage (Alpaca) that executes real equity trades

## Core Flows

### 1. Deposit Flow

#### 1.1 USDC Deposit
```
User → RWA Market → MM → Anchored → Broker
```

**Detailed steps:**
1. The user deposits USDC into the RWA Market
2. The RWA Market deposits USDC into the MM account
3. The MM deposits USDC into the Anchored account
4. Anchored deposits USDC into the Broker account

#### 1.2 XXXX.anc Deposit
```
User → RWA Market → User Account (direct credit)
```

**Detailed steps:**
1. The user deposits XXXX.anc assets into the RWA Market
2. The RWA Market credits the assets to the user account
3. The system mints USD.anc tokens to the user

### 2. Trading Flow

#### 2.1 During Trading Hours
```
Trader → RWA Market → Anchored → Broker → Equity Execution → Token Minting
```

**Detailed steps:**
1. The trader submits a token order to the RWA Market
2. The RWA Market submits the token order to Anchored via the Issuer
3. Anchored submits the equity order to the Broker
4. The Broker executes the equity order
5. Anchored mints the corresponding tokens
6. The RWA Market confirms order completion

#### 2.2 Outside Trading Hours
```
MM → RWA Market (liquidity provision) → Trader (direct fill)
```

**Detailed steps:**
1. The MM provides liquidity to the RWA Market
2. The trader submits a token order to the RWA Market
3. The RWA Market completes the trade directly using MM-provided liquidity

### 3. Withdrawal Flow

#### 3.1 XXXX.anc Withdrawal
```
Trader → RWA Market → Trader (direct return)
```

#### 3.2 USDC Withdrawal
```
Trader → RWA Market → Anchored → Broker → USDC Transfer
```

## Technical Architecture

### System Components

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   Frontend      │    │   RWA Backend   │    │   Alpaca API    │
│   (Web/Mobile)  │◄──►│   (Go Service)  │◄──►│   (Trading)     │
└─────────────────┘    └─────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   Database      │
                       │   (PostgreSQL)  │
                       └─────────────────┘
```

### Core Modules

1. **Trading Module**
   - Order management
   - Trade execution
   - Order status listening

2. **Asset Management**
   - Token minting/burning
   - Asset mapping
   - Balance management

3. **Event Listener**
   - Blockchain event subscription
   - Order status change monitoring
   - Market data listening

4. **Integration Module**
   - Alpaca API integration
   - Blockchain integration
   - External services integration

## Data Flow

### Order Lifecycle

```
Order Creation → Order Validation → Order Submission → Order Execution → Order Completion
    ↓               ↓                 ↓                 ↓                 ↓
 Database       Risk Checks        Alpaca API        Status Updates     Result Notification
```

### Event-Driven Architecture

```
Blockchain Events → Event Listeners → Business Processors → State Updates → Notifications
```