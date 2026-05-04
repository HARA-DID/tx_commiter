const Redis = require('ioredis');
const { Wallet, keccak256, solidityPacked, getBytes, JsonRpcProvider, AbiCoder, zeroPadValue } = require('ethers');

const redis = new Redis("redis://localhost:6379/0");
const streamName = "did-events";

const RPC_URL = "http://70.153.192.125:5625"; 
const CHAIN_ID = 1212;
const DID_FACTORY_ADDRESS = "0xF5fdE394E8446d3b880767dA546B9c33d91bC7B2";

function encodeAddKey(didIndex, publicKey, purpose, keyType) {
    const abiCoder = AbiCoder.defaultAbiCoder();
    const keyData = keccak256(solidityPacked(["address"], [publicKey]));

    const encodedArgs = abiCoder.encode(
        ["uint256", "bytes32", "string", "uint8", "uint8"],
        [didIndex, keyData, "did-root", purpose, keyType]
    );

    const txType = 8;
    const keyIdentifier = "did-root";
    
    const methodId = keccak256(Buffer.from("callExternalDID(uint8,bytes,string)")).substring(0, 10);
    const params = abiCoder.encode(
        ["uint8", "bytes", "string"],
        [txType, encodedArgs, keyIdentifier]
    );

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

async function publishEvent(targetWalletAddress, ownerPrivateKey, didIndex, newPublicKey) {
    const provider = new JsonRpcProvider(RPC_URL);
    const ownerWallet = new Wallet(ownerPrivateKey);
    
    const currentBlockNum = await provider.getBlockNumber();
    const currentBlock = (currentBlockNum + 100).toString(); 
    const encodedData = encodeAddKey(didIndex, newPublicKey, 3, 1); 

    console.log(`[INIT] Smart Wallet: ${targetWalletAddress}`);
    console.log(`[INIT] Signing as Owner: ${ownerWallet.address}`);

    const hash = calculateWalletHash(
        targetWalletAddress,
        DID_FACTORY_ADDRESS,
        0,
        encodedData,
        currentBlock,
        "2"
    );

    console.log(`DEBUG: Data Hash: ${keccak256(encodedData)}`);
    console.log(`DEBUG: Block: ${currentBlock}`);

    const signature = await ownerWallet.signMessage(getBytes(hash));

    const event = {
        id: `evt_add_key_${Date.now()}`,
        type: "ADD_KEY",
        payload: {
            target_address: targetWalletAddress,
            did_index: parseInt(didIndex),
            public_key: newPublicKey,
            key_type: 1,
            purpose: 3,
            signature: Buffer.from(getBytes(signature)).toString('base64'),
            key_identifier: "did-root",
            key_identifier_dst: "did-root",
            client_block_number: currentBlock,
            user_nonce: "2"
        }
    };
    
    try {
        const messageId = await redis.xadd(streamName, '*', 'data', JSON.stringify(event));
        console.log(`✅ Success! Redis Message ID: ${messageId}`);
    } catch (error) {
        console.error(`❌ Error:`, error);
    }
}

async function run() {
    const args = process.argv.slice(2);
    if (args.length < 4) {
        console.error("Usage: node publish_add_key.js <0xSmartWallet> <0xOwnerPrivateKey> <didIndex> <0xNewPublicKey>");
        process.exit(1);
    }
    await publishEvent(args[0], args[1], args[2], args[3]);
    redis.quit();
}

run();

