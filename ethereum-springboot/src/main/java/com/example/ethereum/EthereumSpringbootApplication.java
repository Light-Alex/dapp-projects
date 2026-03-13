package com.example.ethereum;

import org.springframework.boot.SpringApplication;
import org.springframework.boot.autoconfigure.SpringBootApplication;
import org.springframework.scheduling.annotation.EnableScheduling;

@SpringBootApplication
@EnableScheduling
public class EthereumSpringbootApplication {

    public static void main(String[] args) {
        SpringApplication.run(EthereumSpringbootApplication.class, args);
    }
}
