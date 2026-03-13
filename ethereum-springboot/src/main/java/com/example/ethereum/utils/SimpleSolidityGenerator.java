package com.example.ethereum.utils;

import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.nio.file.Files;
import java.nio.file.Paths;

public class SimpleSolidityGenerator {

    public static void main(String[] args) throws Exception {
        try {
            // 检查合约文件是否存在
            File contractFile = new File("src/main/resources/contracts/SimpleStorage.sol");
            if (!contractFile.exists()) {
                System.err.println("错误: 找不到合约文件: " + contractFile.getAbsolutePath());
                return;
            }
            
            // 创建输出目录
            File buildDir = new File("build");
            if (!buildDir.exists()) {
                buildDir.mkdir();
            }
            
            // 确保输出目录存在
            new File("src/main/java/com/example/ethereum/contract/generated").mkdirs();
            
            // 检查solc是否安装
            if (!isCommandAvailable("solc --version") && !isCommandAvailable("solcjs --version")) {
                System.err.println("错误: 未找到solc或solcjs。请安装Solidity编译器。");
                return;
            }
            
            // 步骤1: 编译合约
            System.out.println("正在编译Solidity合约...");
            compileContract(contractFile.getAbsolutePath());
            
            // 步骤2: 检查编译结果
            File[] abiFiles = buildDir.listFiles((dir, name) -> name.endsWith(".abi"));
            File[] binFiles = buildDir.listFiles((dir, name) -> name.endsWith(".bin"));
            
            if (abiFiles == null || abiFiles.length == 0 || binFiles == null || binFiles.length == 0) {
                System.err.println("错误: 未找到编译后的ABI或BIN文件。");
                return;
            }
            
            // 步骤3: 使用web3j-cli生成Java类
            System.out.println("正在生成Java合约类...");
            generateJavaClass(abiFiles[0].getAbsolutePath(), binFiles[0].getAbsolutePath());
            
            System.out.println("操作完成！请检查src/main/java/com/example/ethereum/contract/generated目录。");
            
        } catch (Exception e) {
            System.err.println("发生错误: " + e.getMessage());
            e.printStackTrace();
        }
    }
    
    private static boolean isCommandAvailable(String command) {
        try {
            ProcessBuilder pb;
            if (isWindows()) {
                pb = new ProcessBuilder("cmd", "/c", command);
            } else {
                pb = new ProcessBuilder("sh", "-c", command);
            }
            
            Process process = pb.start();
            int exitCode = process.waitFor();
            return exitCode == 0;
        } catch (Exception e) {
            return false;
        }
    }
    
    private static void compileContract(String contractPath) throws Exception {
        ProcessBuilder pb;
        
        // 使用solc或solcjs
        if (isCommandAvailable("solc --version")) {
            if (isWindows()) {
                pb = new ProcessBuilder("cmd", "/c", 
                        "solc \"" + contractPath + "\" --bin --abi --optimize -o build/");
            } else {
                pb = new ProcessBuilder("solc", 
                        contractPath, "--bin", "--abi", "--optimize", "-o", "build/");
            }
        } else {
            // 使用solcjs
            if (isWindows()) {
                pb = new ProcessBuilder("cmd", "/c", 
                        "solcjs --bin --abi \"" + contractPath + "\" -o build/");
            } else {
                pb = new ProcessBuilder("solcjs", 
                        "--bin", "--abi", contractPath, "-o", "build/");
            }
        }
        
        pb.redirectErrorStream(true);
        Process process = pb.start();
        
        // 显示输出信息
        try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
            String line;
            while ((line = reader.readLine()) != null) {
                System.out.println(line);
            }
        }
        
        int exitCode = process.waitFor();
        if (exitCode != 0) {
            throw new RuntimeException("合约编译失败，退出码: " + exitCode);
        }
    }
    
    private static void generateJavaClass(String abiPath, String binPath) throws Exception {
        ProcessBuilder pb;
        
        // 检查web3j命令是否可用
        if (isCommandAvailable("web3j version") || isCommandAvailable("web3j-cli version")) {
            String command = isCommandAvailable("web3j version") ? "web3j" : "web3j-cli";
            
            if (isWindows()) {
                pb = new ProcessBuilder("cmd", "/c", 
                        command + " generate solidity " +
                        "-a \"" + abiPath + "\" " +
                        "-b \"" + binPath + "\" " +
                        "-o src/main/java " +
                        "-p com.example.ethereum.contract.generated");
            } else {
                pb = new ProcessBuilder(command, "generate", "solidity",
                        "-a", abiPath,
                        "-b", binPath,
                        "-o", "src/main/java",
                        "-p", "com.example.ethereum.contract.generated");
            }
            
            pb.redirectErrorStream(true);
            Process process = pb.start();
            
            // 显示输出信息
            try (BufferedReader reader = new BufferedReader(new InputStreamReader(process.getInputStream()))) {
                String line;
                while ((line = reader.readLine()) != null) {
                    System.out.println(line);
                }
            }
            
            int exitCode = process.waitFor();
            if (exitCode != 0) {
                throw new RuntimeException("Java类生成失败，退出码: " + exitCode);
            }
        } else {
            System.err.println("警告: 未找到web3j命令行工具。请安装web3j-cli。");
            System.err.println("安装指南: https://docs.web3j.io/4.8.7/command_line_tools/");
        }
    }
    
    private static boolean isWindows() {
        return System.getProperty("os.name").toLowerCase().contains("win");
    }
}
