package com.example.ethereum.controller;

import com.example.ethereum.service.ContractService;
import lombok.RequiredArgsConstructor;
import org.springframework.http.ResponseEntity;
import org.springframework.web.bind.annotation.*;
import org.web3j.protocol.core.methods.response.TransactionReceipt;

import java.math.BigInteger;
import java.util.Map;
import java.util.concurrent.CompletableFuture;

@RestController
@RequestMapping("/api/contract")
@RequiredArgsConstructor
public class ContractController {

    private final ContractService contractService;
    
    @GetMapping("/value")
    public ResponseEntity<Map<String, Object>> getValue() {
        BigInteger value = contractService.getValue();
        return ResponseEntity.ok(Map.of(
                "value", value,
                "valueAsString", value.toString()
        ));
    }
    
    @PostMapping("/value")
    public CompletableFuture<ResponseEntity<Map<String, Object>>> setValue(@RequestBody Map<String, String> request) {
        BigInteger newValue = new BigInteger(request.get("value"));
        return contractService.setValue(newValue)
                .thenApply(receipt -> ResponseEntity.ok(Map.of(
                        "transactionHash", receipt.getTransactionHash(),
                        "blockNumber", receipt.getBlockNumber(),
                        "gasUsed", receipt.getGasUsed(),
                        "status", receipt.getStatus()
                )));
    }
    
    /**
     * 部署新的SimpleStorage合约
     * @return 新合约的地址
     */
    @PostMapping("/deploy")
    public ResponseEntity<Map<String, String>> deployContract() {
        String contractAddress = contractService.deployContract();
        return ResponseEntity.ok(Map.of(
                "contractAddress", contractAddress,
                "message", "合约部署成功"
        ));
    }
}
