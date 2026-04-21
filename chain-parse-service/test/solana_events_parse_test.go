package main

import (
	"encoding/base64"
	"strings"
	"testing"
)

type SolanaEvent struct {
	Data      []byte // 解码后的事件数据（含 discriminator）
	ProgramID string // 发出该事件的程序 ID（base58）
}

func Test_SolanaEventsParse(t *testing.T) {

	logMessages := []string{
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program ComputeBudget111111111111111111111111111111 invoke [1]",
		"Program ComputeBudget111111111111111111111111111111 success",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 invoke [1]",
		"Program log: Instruction: ValidateNonce",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 consumed 3372 of 239700 compute units",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 success",
		"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL invoke [1]",
		"Program log: CreateIdempotent",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb invoke [2]",
		"Program log: Instruction: GetAccountDataSize",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb consumed 1444 of 231064 compute units",
		"Program return: TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb qgAAAAAAAAA=",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb success",
		"Program 11111111111111111111111111111111 invoke [2]",
		"Program 11111111111111111111111111111111 success",
		"Program log: Initialize the associated token account",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb invoke [2]",
		"Program log: Instruction: InitializeImmutableOwner",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb consumed 674 of 224798 compute units",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb success",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb invoke [2]",
		"Program log: Instruction: InitializeAccount3",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb consumed 2027 of 221789 compute units",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb success",
		"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL consumed 16849 of 236328 compute units",
		"Program ATokenGPvbdGVxr1b2hvZbsiqW5xWH25efTNsLJA8knL success",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 invoke [1]",
		"Program log: Instruction: BuyExactInPumpFunV3",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P invoke [2]",
		"Program log: Instruction: BuyExactSolIn",
		"Program pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ invoke [3]",
		"Program log: Instruction: GetFees",
		"Program pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ consumed 3136 of 168343 compute units",
		"Program return: pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ AAAAAAAAAABfAAAAAAAAAB4AAAAAAAAA",
		"Program pfeeUxB6jkeY1Hxd7CsFCAjcbHA9rWtchMGdZ6VojVZ success",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb invoke [3]",
		"Program log: Instruction: TransferChecked",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb consumed 2475 of 160497 compute units",
		"Program TokenzQdBNbLqP5VEhdkAS6EPFLC1PHnBqCXEpPxuEb success",
		"Program 11111111111111111111111111111111 invoke [3]",
		"Program 11111111111111111111111111111111 success",
		"Program 11111111111111111111111111111111 invoke [3]",
		"Program 11111111111111111111111111111111 success",
		"Program 11111111111111111111111111111111 invoke [3]",
		"Program 11111111111111111111111111111111 success",
		"Program data: vdt/007mYe4dGrPt6OFkOXJLs/aOoEOX7RLs6KKEjOubEFcitskwT+UZAQEAAAAAr5oxVHcAAAABxc/vyd/tlaoaxur5ahQ+fo0mxDVHgOGRHcky79uRpUENbddpAAAAANnjv0AHAAAAU1SYlk9dAwAC9AzdAAAAAFO8hUq+XgIA6JMUH7GOnxV02BDheOGeMGBOMXWqLkoy38hgByfRBwlfAAAAAAAAAEZxAgAAAAAAvNCU9woZvHz5LwHMkYwhcCpyDrOBWVZnh4V12TB6VMAeAAAAAAAAAHXFAAAAAAAAAQAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAEAAAAGJ1eV9leGFjdF9zb2xfaW4BAAAAAAAAAAAAAAAAAAAAAA==",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P invoke [3]",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P consumed 2060 of 143653 compute units",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P consumed 67334 of 207321 compute units",
		"Program 6EF8rrecthR5Dkzon8Nwu78hRvfCKubJ14M5uBEwF6P success",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 consumed 80307 of 219479 compute units",
		"Program term9YPb9mzAsABaqN71A4xdbxHmpBNZavpBiQKZzN3 success",
		"Program 11111111111111111111111111111111 invoke [1]",
		"Program 11111111111111111111111111111111 success",
		"Program 11111111111111111111111111111111 invoke [1]",
		"Program 11111111111111111111111111111111 success",
	}

	// 追踪当前执行的程序 ID，使用栈结构支持嵌套调用（如 CPI 调用）
	// CPI: Solana 区块链中的 跨程序调用（Cross-Program Invocation）
	var programStack []string
	var events []SolanaEvent

	const invokePrefix = "Program "
	const invokeSuffix = " invoke ["
	const dataPrefix = "Program data: "

	for _, line := range logMessages {
		// 检测程序调用：Program <ID> invoke [<depth>]
		if strings.HasPrefix(line, invokePrefix) && strings.Contains(line, invokeSuffix) {
			trimmed := strings.TrimPrefix(line, invokePrefix)
			idx := strings.Index(trimmed, invokeSuffix)
			if idx > 0 {
				programID := strings.TrimSpace(trimmed[:idx])
				programStack = append(programStack, programID)
			}
			continue
		}

		// 检测程序返回：Program <ID> success / failed
		if strings.HasPrefix(line, invokePrefix) &&
			(strings.HasSuffix(line, " success") || strings.HasSuffix(line, " failed")) {
			if len(programStack) > 0 {
				programStack = programStack[:len(programStack)-1]
			}
			continue
		}

		// 检测事件数据：Program data: <base64>
		if strings.HasPrefix(line, dataPrefix) {
			b64Data := strings.TrimPrefix(line, dataPrefix)
			decoded, err := base64.StdEncoding.DecodeString(b64Data)
			if err != nil {
				continue
			}
			if len(decoded) < 8 {
				continue
			}

			// 取栈顶程序 ID 作为事件来源
			programID := ""
			t.Logf("programStack: %v", programStack)
			if len(programStack) > 0 {
				programID = programStack[len(programStack)-1]
			}

			events = append(events, SolanaEvent{
				Data:      decoded,
				ProgramID: programID,
			})
		}
	}

	// 打印解析后的所有事件信息
	for _, event := range events {
		t.Logf("programID: %v", event.ProgramID)
		t.Logf("data: %v", event.Data)
		t.Log("=========================================================")
	}
}
