import type { ContractParser, Address } from '@derivation-tech/viem-kit';
import { formatAddress, formatWad } from './format';
import type { ParserDeps } from './types';

function stringifyArgs(values: readonly unknown[]): string {
    return values.map((value) => String(value)).join(', ');
}
export function createPocTokenParser(abi:any, deps?: ParserDeps): ContractParser {
    return {
        abi: abi as any,

        async parseTransaction({ functionName, args }) {
            switch (functionName) {
                case 'initialize': {
                    const [name, symbol] = args as readonly [string, string];
                    return `initialize(name: ${name}, symbol: ${symbol})`;
                }
                case 'setOperator': {
                    const [operator] = args as readonly [Address];
                    return `setOperator(operator: ${await formatAddress(operator, deps)})`;
                }
                case 'mint': {
                    const [to, amount] = args as readonly [Address, bigint];
                    return `mint(to: ${await formatAddress(to, deps)}, amount: ${formatWad(amount)})`;
                }
                case 'burn': {
                    const [amount] = args as readonly [bigint];
                    return `burn(amount: ${formatWad(amount)})`;
                }
                case 'burnFrom': {
                    const [from, amount] = args as readonly [Address, bigint];
                    return `burnFrom(from: ${await formatAddress(from, deps)}, amount: ${formatWad(amount)})`;
                }
                default:
                    return `${functionName}(${args.map((v) => String(v)).join(', ')})`;
            }
        },

        async parseEvent(event) {
            const { eventName, args } = event as { eventName: string; args: Record<string, any> };
            switch (eventName) {
                case 'TokensMinted': {
                    const { to, amount } = args as { to: Address; amount: bigint };
                    return `TokensMinted(to: ${await formatAddress(to, deps)}, amount: ${formatWad(amount)})`;
                }
                case 'TokensBurned': {
                    const { from, amount } = args as { from: Address; amount: bigint };
                    return `TokensBurned(from: ${await formatAddress(from, deps)}, amount: ${formatWad(amount)})`;
                }
                case 'Approval': {
                    const { owner, spender, value } = args as { owner: Address; spender: Address; value: bigint };
                    return `Approval(owner: ${await formatAddress(owner, deps)}, spender: ${await formatAddress(spender, deps)}, value: ${formatWad(value)})`;
                }
                case 'Transfer': {
                    const { from, to, value } = args as { from: Address; to: Address; value: bigint };
                    return `Transfer(from: ${await formatAddress(from, deps)}, to: ${await formatAddress(to, deps)}, value: ${formatWad(value)})`;
                }
                default:
                    return `${eventName}(${stringifyArgs(Object.values(args))})`;
            }
        },

        async parseError(error) {
            if ((error as any).name) return (error as any).name;
            if ((error as any).signature) return (error as any).signature;
            return 'PocToken error';
        },
    };
}