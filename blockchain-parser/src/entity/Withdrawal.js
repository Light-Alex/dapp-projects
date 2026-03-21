// CREATE TABLE withdrawal (
//     id SERIAL PRIMARY KEY,
//     account_id INTEGER NOT NULL REFERENCES account(id),
//     amount DECIMAL(36, 18) NOT NULL,
//     token_symbol VARCHAR(10) NOT NULL,
//     to_address VARCHAR(128) NOT NULL,
//     tx_hash VARCHAR(66),
//     status VARCHAR(20) DEFAULT 'init', -- init, processing, success, failed
//     created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );

// CREATE INDEX idx_withdrawal_status ON withdrawal(status);
// CREATE INDEX idx_withdrawal_account ON withdrawal(account_id);


const { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, UpdateDateColumn, Index,  ManyToOne, JoinColumn } = require('typeorm');


const Account = require('./Account');

@Entity('withdrawal')
@Index('idx_withdrawal_status', ['status'])
@Index('idx_withdrawal_account', ['account_id'])
class Withdrawal {
  @PrimaryGeneratedColumn()
  id;

  @ManyToOne(() => Account)
  @JoinColumn({ name: 'account_id' })
  account;

  @Column({ type: 'integer' })
  account_id;

  @Column({ type: 'decimal', precision: 36, scale: 18, default: 0 })
  amount;

  @Column({ type: 'integer', default: 18 })
  token_decimals;

  @Column({ type: 'varchar', length: 10 })
  token_symbol;

  @Column({ type: 'varchar', length: 128 })
  to_address;

  @Column({ type: 'varchar', length: 66, nullable: true })
  tx_hash;

  @Column({ type: 'varchar', length: 20, default: 'init' })
  status;

  @CreateDateColumn()
  created_time;

  @UpdateDateColumn()
  updated_time;
}

module.exports = Withdrawal;
