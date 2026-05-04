const Redis = require('ioredis');
const { Wallet, keccak256, solidityPacked, getBytes, JsonRpcProvider, AbiCoder } = require('ethers');

const redis = new Redis("redis://localhost:6379/0");
const streamName = "did-events";

const RPC_URL = "http://70.153.192.125:5625"; 
const CHAIN_ID = 1212;
const DID_FACTORY_ADDRESS = "0xF5fdE394E8446d3b880767dA546B9c33d91bC7B2"; 

function encodeAddMember(orgDIDIndex, userDID, role) {
    const abiCoder = AbiCoder.defaultAbiCoder();
    const userDIDHash = keccak256(solidityPacked(["string"], [userDID]));
    const encodedData = abiCoder.encode(["bytes32", "uint8"], [userDIDHash, role]);
    
    const txType = 4;
    const methodId = keccak256(Buffer.from("callExternalOrg(uint8,bytes,uint256)")).substring(0, 10);
    const params = abiCoder.encode(
        ["uint8", "bytes", "uint256"],
        [txType, encodedData, orgDIDIndex]
    );

    return {
        callData: methodId + params.substring(2),
        innerData: encodedData
    };
}

function calculateWalletHash(walletAddress, target, value, data, clientBlockNumber, userNonce) {
    const dataHash = keccak256(data);
    const packedData = solidityPacked(
        ['uint256', 'address', 'address', 'uint256', 'bytes32', 'uint256', 'uint256'],
        [CHAIN_ID, walletAddress, target, value, dataHash, clientBlockNumber, userNonce]
    );
    return keccak256(packedData);
}

async function publishEvent(targetWalletAddress, ownerPrivateKey, orgDIDIndex, userDID, role, nonce = "0") {
    const provider = new JsonRpcProvider(RPC_URL);
    const ownerWallet = new Wallet(ownerPrivateKey);
    const currentBlockNum = await provider.getBlockNumber();
    const currentBlock = (currentBlockNum + 100).toString(); 

    const { callData, innerData } = encodeAddMember(orgDIDIndex, userDID, role);

    const hash = calculateWalletHash(targetWalletAddress, DID_FACTORY_ADDRESS, 0, callData, currentBlock, nonce);
    const signature = await ownerWallet.signMessage(getBytes(hash));

    const event = {
        id: `evt_add_member_${Date.now()}`,
        type: "ADD_MEMBER",
        payload: {
            target_address: targetWalletAddress,
            org_did_index: parseInt(orgDIDIndex),
            data: Buffer.from(getBytes(innerData)).toString('base64'),
            signature: Buffer.from(getBytes(signature)).toString('base64'),
            client_block_number: currentBlock,
            user_nonce: nonce,
            multiple_rpc_calls: false
        }
    };
    
    try {
        const messageId = await redis.xadd(streamName, '*', 'data', JSON.stringify(event));
        console.log(`✅ Success! ADD_MEMBER Redis Message ID: ${messageId}`);
        console.log(`DEBUG: Data Hash: ${keccak256(callData)}`);
        console.log(`DEBUG: Nonce: ${nonce}`);
    } catch (error) {
        console.error(`❌ Error:`, error);
    }
}

async function run() {
    const args = process.argv.slice(2);
    if (args.length < 5) {
        console.error("Usage: node 5_add_member.js <SmartWallet> <OwnerKey> <OrgDIDIndex> <UserDID> <Role> [nonce]");
        process.exit(1);
    }
    await publishEvent(args[0], args[1], args[2], args[3], args[4], args[5] || "0");
    redis.quit();
}

run();
