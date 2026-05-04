const Redis = require('ioredis');
const { Wallet, keccak256, solidityPacked, getBytes, JsonRpcProvider, AbiCoder } = require('ethers');

const redis = new Redis("redis://localhost:6379/0");
const streamName = "did-events";

const RPC_URL = "http://70.153.192.125:5625"; 
const CHAIN_ID = 1212;
const DID_FACTORY_ADDRESS = "0xF5fdE394E8446d3b880767dA546B9c33d91bC7B2";

function encodeAddClaim(didIndex, topic, data, uri, signature) {
    const abiCoder = AbiCoder.defaultAbiCoder();
    const encodedArgs = abiCoder.encode(
        ["uint256", "uint8", "bytes", "string", "bytes"],
        [didIndex, topic, data, uri, signature]
    );
    const txType = 10;
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

async function publishEvent(targetWalletAddress, ownerPrivateKey, didIndex, topic, claimData, uri) {
    const provider = new JsonRpcProvider(RPC_URL);
    const ownerWallet = new Wallet(ownerPrivateKey);
    const currentBlockNum = await provider.getBlockNumber();
    const currentBlock = (currentBlockNum + 100).toString(); 
    const dummySignature = "0x" + "00".repeat(65);
    const encodedData = encodeAddClaim(didIndex, topic, claimData, uri, dummySignature);

    const hash = calculateWalletHash(
        targetWalletAddress,
        DID_FACTORY_ADDRESS,
        0,
        encodedData,
        currentBlock,
        "0"
    );

    const signature = await ownerWallet.signMessage(getBytes(hash));

    const event = {
        id: `evt_add_claim_${Date.now()}`,
        type: "ADD_CLAIM",
        payload: {
            target_address: targetWalletAddress,
            did_index: parseInt(didIndex),
            topic: parseInt(topic),
            issuer_address: ownerWallet.address,
            data: Buffer.from(getBytes(claimData)).toString('base64'),
            uri: uri,
            signature: Buffer.from(getBytes(signature)).toString('base64'),
            key_identifier: "did-root",
            client_block_number: currentBlock,
            user_nonce: "0"
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
    if (args.length < 6) {
        console.error("Usage: node publish_add_claim.js <0xSmartWallet> <0xOwnerPrivateKey> <didIndex> <topic> <0xClaimData> <uri>");
        process.exit(1);
    }
    await publishEvent(args[0], args[1], args[2], args[3], args[4], args[5]);
    redis.quit();
}

run();

