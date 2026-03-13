package com.example.ethereum.service;

import com.example.ethereum.config.EthereumConfig;
import com.example.ethereum.contract.generated.SimpleStorage;
import lombok.RequiredArgsConstructor;
import lombok.extern.slf4j.Slf4j;
import org.springframework.stereotype.Service;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.methods.response.TransactionReceipt;
import org.web3j.tx.gas.ContractGasProvider;

import javax.annotation.PostConstruct;
import java.math.BigInteger;
import java.util.concurrent.CompletableFuture;

@Service
@RequiredArgsConstructor
@Slf4j
public class ContractService {

    private final Web3j web3j;
    private final Credentials credentials;
    private final ContractGasProvider gasProvider;
    private final EthereumConfig ethereumConfig;
    
    private SimpleStorage contract;
    
    @PostConstruct
    public void init() {
        contract = SimpleStorage.load(
                ethereumConfig.getContractAddress(),
                web3j,
                credentials,
                gasProvider
        );
        log.info("Contract initialized with address: {}", ethereumConfig.getContractAddress());
    }
    
    public BigInteger getValue() {
        try {
            return contract.getValue().send();
        } catch (Exception e) {
            log.error("Error getting value from contract", e);
            throw new RuntimeException("Failed to get value from contract", e);
        }
    }
    
    public CompletableFuture<TransactionReceipt> setValue(BigInteger newValue) {
        log.info("Setting value to: {}", newValue);
        return contract.setValue(newValue).sendAsync()
                .thenApply(receipt -> {
                    log.info("Transaction completed. Hash: {}", receipt.getTransactionHash());
                    return receipt;
                })
                .exceptionally(e -> {
                    log.error("Error setting value", e);
                    throw new RuntimeException("Failed to set value", e);
                });
    }
    
    /**
     * 部署新的SimpleStorage合约
     * @return 部署的合约地址
     */
    public String deployContract() {
        try {
            log.info("开始部署新的SimpleStorage合约...");
            SimpleStorage newContract = SimpleStorage.deploy(
                    web3j,
                    credentials,
                    gasProvider
            ).send();
            
            String contractAddress = newContract.getContractAddress();
            log.info("合约部署成功! 合约地址: {}", contractAddress);
            return contractAddress;
        } catch (Exception e) {
            log.error("部署合约失败", e);
            throw new RuntimeException("部署合约失败", e);
        }
    }
}
