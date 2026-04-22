import { formatUnits } from 'viem';
import type { Address } from '@derivation-tech/viem-kit';
import type { ParserDeps } from './types.js';

export const DEFAULT_DECIMALS = 18;

function shortenDecimals(value: string, fixedDecimals = 6): string {
    if (!value.includes('.')) return value;
    const [_integer, fraction] = value.split('.');
    if (fraction.length <= fixedDecimals) return value;
    const trimmed = Number(value).toFixed(fixedDecimals);
    return trimmed;
}

function toBigInt(value: bigint | number | string): bigint {
    if (typeof value === 'bigint') return value;
    if (typeof value === 'number') return BigInt(Math.trunc(value));
    if (typeof value === 'string' && (value.startsWith('0x') || value.startsWith('0X'))) {
        return BigInt(value);
    }
    return BigInt(value as any);
}

export function formatUnitsSafe(value: bigint | number | string, decimals: number, fixedDecimals = 6): string {
    const formatted = formatUnits(toBigInt(value), decimals);
    return shortenDecimals(formatted, fixedDecimals);
}

export function formatWad(value: bigint | number | string, fixedDecimals = 6): string {
    return formatUnitsSafe(value, DEFAULT_DECIMALS, fixedDecimals);
}

export function formatTimestamp(value: bigint | number | string): string {
    const seconds = Number(toBigInt(value));
    if (seconds <= 0) return seconds.toString();
    return new Date(seconds * 1000).toISOString();
}

export async function formatAddress(address: Address, deps?: ParserDeps): Promise<string> {
    if (deps?.resolveAddress) {
        const name = await deps.resolveAddress(address);
        if (name && name !== 'UNKNOWN' && name !== address) {
            return `[${name}]${address}`;
        }
    }
    return address;
}