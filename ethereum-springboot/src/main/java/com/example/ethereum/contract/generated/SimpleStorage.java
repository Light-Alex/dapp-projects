package com.example.ethereum.contract.generated;

import io.reactivex.Flowable;
import java.math.BigInteger;
import java.util.ArrayList;
import java.util.Arrays;
import java.util.Collections;
import java.util.List;
import org.web3j.abi.EventEncoder;
import org.web3j.abi.TypeReference;
import org.web3j.abi.datatypes.Address;
import org.web3j.abi.datatypes.Event;
import org.web3j.abi.datatypes.Function;
import org.web3j.abi.datatypes.Type;
import org.web3j.abi.datatypes.generated.Uint256;
import org.web3j.crypto.Credentials;
import org.web3j.protocol.Web3j;
import org.web3j.protocol.core.DefaultBlockParameter;
import org.web3j.protocol.core.RemoteCall;
import org.web3j.protocol.core.RemoteFunctionCall;
import org.web3j.protocol.core.methods.request.EthFilter;
import org.web3j.protocol.core.methods.response.BaseEventResponse;
import org.web3j.protocol.core.methods.response.Log;
import org.web3j.protocol.core.methods.response.TransactionReceipt;
import org.web3j.tx.Contract;
import org.web3j.tx.TransactionManager;
import org.web3j.tx.gas.ContractGasProvider;

/**
 * <p>Auto generated code.
 * <p><strong>Do not modify!</strong>
 * <p>Please use the <a href="https://docs.web3j.io/command_line.html">web3j command line tools</a>,
 * or the org.web3j.codegen.SolidityFunctionWrapperGenerator in the 
 * <a href="https://github.com/LFDT-web3j/web3j/tree/main/codegen">codegen module</a> to update.
 *
 * <p>Generated with web3j version 1.7.0.
 */
@SuppressWarnings("rawtypes")
public class SimpleStorage extends Contract {
    public static final String BINARY = "6080604052348015600e575f5ffd5b506101e28061001c5f395ff3fe608060405234801561000f575f5ffd5b5060043610610034575f3560e01c806320965255146100385780635524107714610056575b5f5ffd5b610040610072565b60405161004d91906100d4565b60405180910390f35b610070600480360381019061006b919061011b565b61007a565b005b5f5f54905090565b805f819055507fc53a6612a7428e1fc89cb87169d69d64eaefe94f5b45d43f80f1ad9d7065d2c881336040516100b1929190610185565b60405180910390a150565b5f819050919050565b6100ce816100bc565b82525050565b5f6020820190506100e75f8301846100c5565b92915050565b5f5ffd5b6100fa816100bc565b8114610104575f5ffd5b50565b5f81359050610115816100f1565b92915050565b5f602082840312156101305761012f6100ed565b5b5f61013d84828501610107565b91505092915050565b5f73ffffffffffffffffffffffffffffffffffffffff82169050919050565b5f61016f82610146565b9050919050565b61017f81610165565b82525050565b5f6040820190506101985f8301856100c5565b6101a56020830184610176565b939250505056fea264697066735822122022a5af0b10edc7600658d0a17ec3d6a03a3c92d265a4c9245434d3dccaffeaa764736f6c63430008220033";

    private static String librariesLinkedBinary;

    public static final String FUNC_GETVALUE = "getValue";

    public static final String FUNC_SETVALUE = "setValue";

    public static final Event VALUECHANGED_EVENT = new Event("ValueChanged", 
            Arrays.<TypeReference<?>>asList(new TypeReference<Uint256>() {}, new TypeReference<Address>() {}));
    ;

    @Deprecated
    protected SimpleStorage(String contractAddress,
            Web3j web3j, Credentials credentials, BigInteger gasPrice, BigInteger gasLimit) {
        super(BINARY, contractAddress, web3j, credentials, gasPrice, gasLimit);
    }

    protected SimpleStorage(String contractAddress,
            Web3j web3j, Credentials credentials, ContractGasProvider contractGasProvider) {
        super(BINARY, contractAddress, web3j, credentials, contractGasProvider);
    }

    @Deprecated
    protected SimpleStorage(String contractAddress,
            Web3j web3j, TransactionManager transactionManager, BigInteger gasPrice,
            BigInteger gasLimit) {
        super(BINARY, contractAddress, web3j, transactionManager, gasPrice, gasLimit);
    }

    protected SimpleStorage(String contractAddress,
            Web3j web3j, TransactionManager transactionManager,
            ContractGasProvider contractGasProvider) {
        super(BINARY, contractAddress, web3j, transactionManager, contractGasProvider);
    }

    public static List<ValueChangedEventResponse> getValueChangedEvents(
            TransactionReceipt transactionReceipt) {
        List<Contract.EventValuesWithLog> valueList = staticExtractEventParametersWithLog(VALUECHANGED_EVENT, transactionReceipt);
        ArrayList<ValueChangedEventResponse> responses = new ArrayList<ValueChangedEventResponse>(valueList.size());
        for (Contract.EventValuesWithLog eventValues : valueList) {
            ValueChangedEventResponse typedResponse = new ValueChangedEventResponse();
            typedResponse.log = eventValues.getLog();
            typedResponse.newValue = (BigInteger) eventValues.getNonIndexedValues().get(0).getValue();
            typedResponse.changedBy = (String) eventValues.getNonIndexedValues().get(1).getValue();
            responses.add(typedResponse);
        }
        return responses;
    }

    public static ValueChangedEventResponse getValueChangedEventFromLog(Log log) {
        Contract.EventValuesWithLog eventValues = staticExtractEventParametersWithLog(VALUECHANGED_EVENT, log);
        ValueChangedEventResponse typedResponse = new ValueChangedEventResponse();
        typedResponse.log = log;
        typedResponse.newValue = (BigInteger) eventValues.getNonIndexedValues().get(0).getValue();
        typedResponse.changedBy = (String) eventValues.getNonIndexedValues().get(1).getValue();
        return typedResponse;
    }

    public Flowable<ValueChangedEventResponse> valueChangedEventFlowable(EthFilter filter) {
        return web3j.ethLogFlowable(filter).map(log -> getValueChangedEventFromLog(log));
    }

    public Flowable<ValueChangedEventResponse> valueChangedEventFlowable(
            DefaultBlockParameter startBlock, DefaultBlockParameter endBlock) {
        EthFilter filter = new EthFilter(startBlock, endBlock, getContractAddress());
        filter.addSingleTopic(EventEncoder.encode(VALUECHANGED_EVENT));
        return valueChangedEventFlowable(filter);
    }

    public RemoteFunctionCall<BigInteger> getValue() {
        final Function function = new Function(FUNC_GETVALUE, 
                Arrays.<Type>asList(), 
                Arrays.<TypeReference<?>>asList(new TypeReference<Uint256>() {}));
        return executeRemoteCallSingleValueReturn(function, BigInteger.class);
    }

    public RemoteFunctionCall<TransactionReceipt> setValue(BigInteger _value) {
        final Function function = new Function(
                FUNC_SETVALUE, 
                Arrays.<Type>asList(new org.web3j.abi.datatypes.generated.Uint256(_value)), 
                Collections.<TypeReference<?>>emptyList());
        return executeRemoteCallTransaction(function);
    }

    @Deprecated
    public static SimpleStorage load(
            String contractAddress, Web3j web3j, Credentials credentials, BigInteger gasPrice,
            BigInteger gasLimit) {
        return new SimpleStorage(contractAddress, web3j, credentials, gasPrice, gasLimit);
    }

    @Deprecated
    public static SimpleStorage load(
            String contractAddress, Web3j web3j, TransactionManager transactionManager,
            BigInteger gasPrice, BigInteger gasLimit) {
        return new SimpleStorage(contractAddress, web3j, transactionManager, gasPrice, gasLimit);
    }

    public static SimpleStorage load(
            String contractAddress, Web3j web3j, Credentials credentials,
            ContractGasProvider contractGasProvider) {
        return new SimpleStorage(contractAddress, web3j, credentials, contractGasProvider);
    }

    public static SimpleStorage load(
            String contractAddress, Web3j web3j, TransactionManager transactionManager,
            ContractGasProvider contractGasProvider) {
        return new SimpleStorage(contractAddress, web3j, transactionManager, contractGasProvider);
    }

    public static RemoteCall<SimpleStorage> deploy(
            Web3j web3j, Credentials credentials, ContractGasProvider contractGasProvider) {
        return deployRemoteCall(SimpleStorage.class, web3j, credentials, contractGasProvider, getDeploymentBinary(), "");
    }

    @Deprecated
    public static RemoteCall<SimpleStorage> deploy(
            Web3j web3j, Credentials credentials, BigInteger gasPrice, BigInteger gasLimit) {
        return deployRemoteCall(SimpleStorage.class, web3j, credentials, gasPrice, gasLimit, getDeploymentBinary(), "");
    }

    public static RemoteCall<SimpleStorage> deploy(
            Web3j web3j, TransactionManager transactionManager,
            ContractGasProvider contractGasProvider) {
        return deployRemoteCall(SimpleStorage.class, web3j, transactionManager, contractGasProvider, getDeploymentBinary(), "");
    }

    @Deprecated
    public static RemoteCall<SimpleStorage> deploy(
            Web3j web3j, TransactionManager transactionManager, BigInteger gasPrice,
            BigInteger gasLimit) {
        return deployRemoteCall(SimpleStorage.class, web3j, transactionManager, gasPrice, gasLimit, getDeploymentBinary(), "");
    }


    private static String getDeploymentBinary() {
        if (librariesLinkedBinary != null) {
            return librariesLinkedBinary;
        } else {
            return BINARY;
        }
    }

    public static class ValueChangedEventResponse extends BaseEventResponse {
        public BigInteger newValue;

        public String changedBy;
    }
}
