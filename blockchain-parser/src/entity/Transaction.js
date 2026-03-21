// CREATE TABLE transaction (
//     id SERIAL PRIMARY KEY,
//     tx_hash VARCHAR(66) UNIQUE NOT NULL,
//     block_number BIGINT NOT NULL,
//     from_address VARCHAR(128) NOT NULL,
//     to_address VARCHAR(128) NOT NULL,
//     value DECIMAL(36, 18) NOT NULL,
//     token_address VARCHAR(128),
//     token_symbol VARCHAR(10),
//     status SMALLINT NOT NULL, -- 0: 失败, 1: 成功
//     created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );

// CREATE INDEX idx_tx_hash ON transaction(tx_hash);
// CREATE INDEX idx_tx_from ON transaction(from_address);
// CREATE INDEX idx_tx_to ON transaction(to_address);
// CREATE INDEX idx_tx_block ON transaction(block_number);


const { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, Index } = require('typeorm');

@Entity('transaction')
@Index('idx_tx_hash', ['tx_hash'])
@Index('idx_tx_from', ['from_address'])
@Index('idx_tx_to', ['to_address'])
@Index('idx_tx_block', ['block_number'])
class Transaction {
  @PrimaryGeneratedColumn()
  id;

  @Column({ type: 'varchar', length: 66, unique: true })
  tx_hash;

  @Column({ type: 'bigint' })
  block_number;

  @Column({ type: 'varchar', length: 128 })
  from_address;

  @Column({ type: 'varchar', length: 128 })
  to_address;

  @Column({ type: 'decimal', precision: 36, scale: 18, default: 0 })
  value;

  @Column({ type: 'integer', default: 18 })
  token_decimals;

  @Column({ type: 'varchar', length: 128, nullable: true })
  token_address;

  @Column({ type: 'varchar', length: 10, nullable: true })
  token_symbol;

  @Column({ type: 'smallint', default: 0 })
  status;

  @CreateDateColumn()
  created_time;

}

module.exports = Transaction;
