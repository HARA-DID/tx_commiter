const Redis = require('ioredis');
const { Wallet, keccak256, solidityPacked, getBytes, JsonRpcProvider, AbiCoder } = require('ethers');

const redis = new Redis("redis://localhost:6379/0");
const streamName = "did-events";

const RPC_URL = "http://70.153.192.125:5625"; 
const CHAIN_ID = 1212;
const DID_FACTORY_ADDRESS = "0xF5fdE394E8446d3b880767dA546B9c33d91bC7B2";

function encodeRemoveOps(didIndex, data, txType) {
    const abiCoder = AbiCoder.defaultAbiCoder();
    const encodedArgs = abiCoder.encode(["uint256", "bytes32"], [didIndex, data]);
    const keyIdentifier = "did-root";
    const methodId = keccak256(Buffer.from("callExternalDID(uint8,bytes,string)")).substring(0, 10);
    const params = abiCoder.encode(["uint8", "bytes", "string"], [txType, encodedArgs, keyIdentifier]);
    return methodId + params.substring(2);
}

function calculateWalletHash(walletAddress, target, value, data, clientBlockNumber, userNonce) {
    const dataHash = keccak256(data);
    const packedData = solidityPacked(
        ['uint256', 'address', 'address', 'uint256', 'bytes32', 'uint256', 'uint256'],
        [CHAIN_ID, walletAddress, target, value, dataHash, clientBlockNumber, userNonce]
    );
    return keccak256(packedData);
}

async function publishEvent(targetWalletAddress, ownerPrivateKey, didIndex, data, txType, eventType, payloadKey) {
    const provider = new JsonRpcProvider(RPC_URL);
    const ownerWallet = new Wallet(ownerPrivateKey);
    const currentBlockNum = await provider.getBlockNumber();
    const currentBlock = (currentBlockNum + 100).toString(); 

    const encodedData = encodeRemoveOps(didIndex, data, txType);
    const hash = calculateWalletHash(targetWalletAddress, DID_FACTORY_ADDRESS, 0, encodedData, currentBlock, "0");
    const signature = await ownerWallet.signMessage(getBytes(hash));

    const event = {
        id: `evt_${eventType.toLowerCase()}_${Date.now()}`,
        type: eventType,
        payload: {
            target_address: targetWalletAddress,
            did_index: parseInt(didIndex),
            [payloadKey]: data,
            signature: Buffer.from(getBytes(signature)).toString('base64'),
            key_identifier: "did-root",
            client_block_number: currentBlock,
            user_nonce: "0"
        }
    };
    
    try {
        const messageId = await redis.xadd(streamName, '*', 'data', JSON.stringify(event));
        console.log(`✅ Success! ${eventType} Redis Message ID: ${messageId}`);
    } catch (error) {
        console.error(`❌ Error:`, error);
    }
}

async function run() {
    const args = process.argv.slice(2);
    if (args.length < 4) {
        console.error("Usage: node publish_remove_did.js <SmartWallet> <OwnerKey> <didIndex> <Hash> <REMOVE_KEY|REMOVE_CLAIM>");
        process.exit(1);
    }
    const type = args[4] === "REMOVE_KEY" ? "REMOVE_KEY" : "REMOVE_CLAIM";
    const txType = type === "REMOVE_KEY" ? 9 : 11;
    const payloadKey = type === "REMOVE_KEY" ? "key_data_hashed" : "claim_id";
    await publishEvent(args[0], args[1], args[2], args[3], txType, type, payloadKey);
    redis.quit();
}

run();

