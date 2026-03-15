// npm i @chainlink/contracts --save
import "@chainlink/contracts/src/v0.8/automation/AutomationCompatible.sol";

contract IntervalCounter is AutomationCompatibleInterface {
    uint256 public counter;
    uint256 public immutable interval;
    uint256 public lastTimeStamp;

    event CounterTick(uint256 counter, uint256 timestamp);

    constructor(uint256 _intervalSeconds) {
        require(_intervalSeconds > 0, "interval must be > 0");
        interval = _intervalSeconds;
        lastTimeStamp = block.timestamp;
    }

    // Automation 节点离线调用，判断是否需要执行
    function checkUpkeep(bytes calldata)
        external
        view
        override
        returns (bool upkeepNeeded, bytes memory performData)
    {
        upkeepNeeded = (block.timestamp - lastTimeStamp) >= interval;
        performData = bytes(""); // 本例无需额外数据
    }

    // Automation 节点链上调用，执行实际任务
    function performUpkeep(bytes calldata) external override {
        if ((block.timestamp - lastTimeStamp) < interval) {
            // 双重检查，防止竞争
            revert("Not time yet");
        }
        lastTimeStamp = block.timestamp;
        counter += 1;
        emit CounterTick(counter, block.timestamp);
    }
}
