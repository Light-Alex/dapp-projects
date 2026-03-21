// CREATE TABLE account (
//     id SERIAL PRIMARY KEY,
//     email VARCHAR(128) UNIQUE NOT NULL,
//     address VARCHAR(128) UNIQUE NOT NULL,
//     bnb_amount DECIMAL(36, 18) DEFAULT 0,
//     usdt_amount DECIMAL(36, 6) DEFAULT 0,
//     created_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
//     updated_time TIMESTAMP DEFAULT CURRENT_TIMESTAMP
// );

// CREATE INDEX idx_account_address ON account(address);


const { Entity, PrimaryGeneratedColumn, Column, CreateDateColumn, UpdateDateColumn, Index } = require('typeorm');

@Entity('account')
@Index('idx_account_address', ['address'])
class Account {
  @PrimaryGeneratedColumn()
  id;

  @Column({ type: 'varchar', length: 128, unique: true })
  email;

  @Column({ type: 'varchar', length: 128, unique: true })
  address;

  @Column({ type: 'decimal', precision: 36, scale: 18, default: 0 })
  bnb_amount;

  @Column({ type: 'decimal', precision: 36, scale: 6, default: 0 })
  usdt_amount;

  @CreateDateColumn()
  created_time;

  @UpdateDateColumn()
  updated_time;
}

module.exports = Account;