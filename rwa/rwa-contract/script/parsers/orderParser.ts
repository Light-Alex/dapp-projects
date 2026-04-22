import type { Abi } from 'viem';
import type { ContractParser, Address } from '@derivation-tech/viem-kit';
import { formatAddress, formatTimestamp, formatWad } from './format';
import type { ParserDeps } from './types';


const SIDE_LABELS = ['Buy', 'Sell'] as const;
const ORDER_TYPE_LABELS = ['Market', 'Limit'] as const;
const TIF_LABELS = ['DAY', 'GTC', 'OPG', 'IOC', 'FOK', 'GTX', 'GTD', 'CLS'] as const;
const STATUS_LABELS = ['Pending', 'Executed', 'CancelRequested', 'Cancelled'] as const;

function formatEnum(labels: readonly string[], value: number | bigint): string {
    const index = Number(value);
    return labels[index] ?? index.toString();
}

function stringifyArgs(values: readonly unknown[]): string {
    return values.map((value) => String(value)).join(', ');
}


export function createOrderParser(abi: any, deps?: ParserDeps): ContractParser {
    return {
        abi: abi as any,
        async parseTransaction({ functionName, args }) {
            switch (functionName) {
                case 'submitOrder': {
                    const [symbol, qty, price, side, orderType, tif] = args as readonly [
                        string,
                        bigint,
                        bigint,
                        number,
                        number,
                        number,
                    ];
                    const payload = [
                        `symbol: ${symbol}`,
                        `qty: ${formatWad(qty)}`,
                        `price: ${formatWad(price)}`,
                        `side: ${formatEnum(SIDE_LABELS, side)}`,
                        `orderType: ${formatEnum(ORDER_TYPE_LABELS, orderType)}`,
                        `tif: ${formatEnum(TIF_LABELS, tif)}`,
                    ];
                    return `submitOrder(${payload.join(', ')})`;
                }
                case 'cancelOrderIntent': {
                    const [orderId] = args as readonly [bigint];
                    return `cancelOrderIntent(orderId: ${orderId})`;
                }
                case 'markExecuted': {
                    const [orderId, refundAmount] = args as readonly [bigint, bigint];
                    return `markExecuted(orderId: ${orderId}, refundAmount: ${formatWad(refundAmount)})`;
                }
                case 'cancelOrder': {
                    const [orderId] = args as readonly [bigint];
                    return `cancelOrder(orderId: ${orderId})`;
                }
                case 'setSymbolToken': {
                    const [symbol, token] = args as readonly [string, Address];
                    return `setSymbolToken(symbol: ${symbol}, token: ${await formatAddress(token, deps)})`;
                }
                case 'setBackend': {
                    const [backend] = args as readonly [Address];
                    return `setBackend(backend: ${await formatAddress(backend, deps)})`;
                }
                case 'initialize': {
                    const [usdm, admin, backend] = args as readonly [Address, Address, Address];
                    const payload = [
                        `usdm: ${await formatAddress(usdm, deps)}`,
                        `admin: ${await formatAddress(admin, deps)}`,
                        `backend: ${await formatAddress(backend, deps)}`,
                    ];
                    return `initialize(${payload.join(', ')})`;
                }
                default: {
                    return `${functionName}(${args.map((v) => String(v)).join(', ')})`;
                }
            }
        },

        async parseEvent(event) {
            const { eventName, args } = event as { eventName: string; args: Record<string, any> };
            switch (eventName) {
                case 'OrderSubmitted': {
                    const { user, orderId, symbol, qty, price, side, orderType, tif, blockTimestamp } = args as {
                        user: Address;
                        orderId: bigint;
                        symbol: string;
                        qty: bigint;
                        price: bigint;
                        side: number;
                        orderType: number;
                        tif: number;
                        blockTimestamp: number | bigint;
                    };
                    const payload = [
                        `user: ${await formatAddress(user, deps)}`,
                        `orderId: ${orderId}`,
                        `symbol: ${symbol}`,
                        `qty: ${formatWad(qty)}`,
                        `price: ${formatWad(price)}`,
                        `side: ${formatEnum(SIDE_LABELS, side)}`,
                        `orderType: ${formatEnum(ORDER_TYPE_LABELS, orderType)}`,
                        `tif: ${formatEnum(TIF_LABELS, tif)}`,
                        `timestamp: ${formatTimestamp(blockTimestamp)}`,
                    ];
                    return `OrderSubmitted(${payload.join(', ')})`;
                }
                case 'CancelRequested': {
                    const { user, orderId, blockTimestamp } = args as {
                        user: Address;
                        orderId: bigint;
                        blockTimestamp: number | bigint;
                    };
                    const payload = [
                        `user: ${await formatAddress(user, deps)}`,
                        `orderId: ${orderId}`,
                        `timestamp: ${formatTimestamp(blockTimestamp)}`,
                    ];
                    return `CancelRequested(${payload.join(', ')})`;
                }
                case 'OrderExecuted': {
                    const { orderId, refundAmount, tif } = args as {
                        orderId: bigint;
                        refundAmount: bigint;
                        tif: number;
                    };
                    const payload = [
                        `orderId: ${orderId}`,
                        `refundAmount: ${formatWad(refundAmount)}`,
                        `tif: ${formatEnum(TIF_LABELS, tif)}`,
                    ];
                    return `OrderExecuted(${payload.join(', ')})`;
                }
                case 'OrderCancelled': {
                    const { orderId, user, asset, refundAmount, side, orderType, tif, previousStatus } = args as {
                        orderId: bigint;
                        user: Address;
                        asset: Address;
                        refundAmount: bigint;
                        side: number;
                        orderType: number;
                        tif: number;
                        previousStatus: number;
                    };
                    const payload = [
                        `orderId: ${orderId}`,
                        `user: ${await formatAddress(user, deps)}`,
                        `asset: ${await formatAddress(asset, deps)}`,
                        `refundAmount: ${formatWad(refundAmount)}`,
                        `side: ${formatEnum(SIDE_LABELS, side)}`,
                        `orderType: ${formatEnum(ORDER_TYPE_LABELS, orderType)}`,
                        `tif: ${formatEnum(TIF_LABELS, tif)}`,
                        `previousStatus: ${formatEnum(STATUS_LABELS, previousStatus)}`,
                    ];
                    return `OrderCancelled(${payload.join(', ')})`;
                }
                default:
                    return `${eventName}(${Object.values(args).map(v => String(v)).join(', ')})`;
            }
        },

        async parseError(error) {
            if ((error as any).name) return (error as any).name;
            if ((error as any).signature) return (error as any).signature;
            return 'OrderContract error';
        },
    };
}