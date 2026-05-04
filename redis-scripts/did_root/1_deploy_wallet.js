const Redis = require('ioredis');
const { Wallet, hexlify, randomBytes } = require('ethers');

// From your .env: REDIS_URL=redis://localhost:6379/0
const redis = new Redis("redis://localhost:6379/0");

const streamName = "did-events";
const wallet = Wallet.createRandom();
const owner = wallet.address;
const privateKey = wallet.privateKey;
async function publishEvent(index) {
    const salt = hexlify(randomBytes(32));

    console.log(`[${index}] Generated Wallet: ${owner}`);
    console.log(`[${index}] Private Key: ${privateKey}`);
    console.log(`[${index}] Salt: ${salt}`);

    const eventId = `evt_deploy_wallet_${Date.now()}_${index}`;
    
    const event = {
        id: eventId,
        type: "DEPLOY_WALLET",
        payload: {
            owner: owner,
            salt: salt,
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
        console.log(`[${index}] ✅ Success! Redis Message ID: ${messageId}`);
    } catch (error) {
        console.error(`[${index}] ❌ Error publishing event:`, error);
    }
}

async function run() {
    console.log("Publishing 100 events...");
    const promises = [];
    for (let i = 1; i < 2; i++) {
        promises.push(publishEvent(i));
    }
    
    await Promise.all(promises);
    
    console.log("Finished. Disconnecting...");
    redis.quit();
}

run();

