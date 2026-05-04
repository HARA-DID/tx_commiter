const Redis = require('ioredis');
const { Wallet, keccak256, solidityPacked, getBytes, JsonRpcProvider, AbiCoder } = require('ethers');

const redis = new Redis("redis://localhost:6379/0");
const streamName = "did-events";

const RPC_URL = "http://70.153.192.125:5625"; 
const CHAIN_ID = 1212;
const DID_FACTORY_ADDRESS = "0xF5fdE394E8446d3b880767dA546B9c33d91bC7B2"; 

function encodeCreateDID(did) {
    const abiCoder = AbiCoder.defaultAbiCoder();
    const encodedDID = abiCoder.encode(["string"], [did]);
    const txType = 1;
    const keyIdentifier = "did-root";
    const methodId = keccak256(Buffer.from("callExternalDID(uint8,bytes,string)")).substring(0, 10);
    
    const params = abiCoder.encode(
        ["uint8", "bytes", "string"],
        [txType, encodedDID, keyIdentifier]
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

async function publishEvent(targetWalletAddress, ownerPrivateKey) {
    const provider = new JsonRpcProvider(RPC_URL);
    const ownerWallet = new Wallet(ownerPrivateKey);
    
    console.log(`[INIT] Smart Wallet: ${targetWalletAddress}`);
    console.log(`[INIT] Signing as Owner: ${ownerWallet.address}`);

    const currentBlockNum = await provider.getBlockNumber();
    const currentBlock = (currentBlockNum + 100).toString(); 
    console.log(`[INFO] Current Block: ${currentBlockNum}, Using: ${currentBlock}`);

    const did = `did:hara:${targetWalletAddress.toLowerCase()}-0`;
    const encodedData = encodeCreateDID(did);

    const userOp = {
        target: DID_FACTORY_ADDRESS,
        value: 0,
        data: encodedData,
        clientBlockNumber: currentBlock,
        userNonce: "0" 
    };

    const hash = calculateWalletHash(
        targetWalletAddress,
        userOp.target,
        userOp.value,
        userOp.data,
        userOp.clientBlockNumber,
        userOp.userNonce
    );

    const signature = await ownerWallet.signMessage(getBytes(hash));

    const eventId = `evt_create_did_${Date.now()}`;
    const event = {
        id: eventId,
        type: "CREATE_DID",
        payload: {
            sender: targetWalletAddress,
            target_address: targetWalletAddress,
            did: did,
            signature: Buffer.from(getBytes(signature)).toString('base64'),
            key_identifier: "did-root",
            client_block_number: currentBlock, 
            user_nonce: userOp.userNonce,       
            multiple_rpc_calls: false
        }
    };
    
    try {
        const messageId = await redis.xadd(
            streamName, 
            '*',
            'data', 
            JSON.stringify(event)
        );
        console.log(`✅ Success! Redis Message ID: ${messageId}`);
        console.log(`DEBUG: Data Hash: ${keccak256(encodedData)}`);
        console.log(`DEBUG: Block: ${currentBlock}`);
    } catch (error) {
        console.error(`❌ Error:`, error);
    }
}

async function run() {
    const args = process.argv.slice(2);
    if (args.length < 2) {
        console.error("Usage: node publish_create_did.js <0xSmartWallet> <0xOwnerPrivateKey>");
        process.exit(1);
    }
    await publishEvent(args[0], args[1]);
    redis.quit();
}

run();

