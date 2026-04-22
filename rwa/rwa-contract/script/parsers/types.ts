import type { Address } from '@derivation-tech/viem-kit';

export interface ParserDeps {
    /** Optional address resolver used to annotate addresses with friendly names */
    resolveAddress?: (address: Address) => string | Promise<string>;

    /** Optional public client for future extensions; not required by current parsers */
    publicClient?: unknown;

    /** Optional native token placeholder address (not used by current parsers) */
    nativeTokenAddress?: Address;
}

export type MaybePromise<T> = T | Promise<T>;