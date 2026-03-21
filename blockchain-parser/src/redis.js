const Redis = require('ioredis');
const config = require('./config');

class RedisService {
  constructor() {
    this.redis = new Redis(config.redis);
  }

  async getLastProcessedBlock() {
    return await this.redis.get('last_processed_block');
  }

  async setLastProcessedBlock(blockNumber) {
    return await this.redis.set('last_processed_block', blockNumber);
  }

  async isTxProcessing(txHash) {
    return await this.redis.get(`tx:processing:${txHash}`);
  }

  async setTxProcessing(txHash, value = '1', expire = 300) {
    return await this.redis.set(`tx:processing:${txHash}`, value, 'EX', expire);
  }

  async removeTxProcessing(txHash) {
    return await this.redis.del(`tx:processing:${txHash}`);
  }

  async cacheUserAddress(address, userId) {
    return await this.redis.set(`address:user:${address}`, userId, 'EX', 3600);
  }

  async getCachedUserByAddress(address) {
    return await this.redis.get(`address:user:${address}`);
  }
}

module.exports = new RedisService();