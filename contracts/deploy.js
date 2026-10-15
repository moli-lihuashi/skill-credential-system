// CredentialRegistry.sol 合约部署脚本
// 使用方法:
//   npm install ethers solc
//   node deploy.js
//
// 部署后将输出的合约地址填入后端环境变量 CONTRACT_ADDR

const fs = require('fs');
const path = require('path');

async function main() {
    // 动态加载依赖（如未安装则提示）
    let ethers, solc;
    try {
        ethers = require('ethers');
        solc = require('solc');
    } catch (e) {
        console.error('请先安装依赖: npm install ethers solc');
        process.exit(1);
    }

    // 读取合约源码
    const source = fs.readFileSync(
        path.join(__dirname, 'CredentialRegistry.sol'),
        'utf8'
    );

    // 编译合约
    console.log('正在编译合约...');
    const input = {
        language: 'Solidity',
        sources: {
            'CredentialRegistry.sol': { content: source }
        },
        settings: {
            optimizer: { enabled: true, runs: 200 },
            viaIR: true,
            outputSelection: {
                '*': { '*': ['abi', 'evm.bytecode'] }
            }
        }
    };

    const output = JSON.parse(solc.compile(JSON.stringify(input)));
    if (output.errors) {
        output.errors.forEach(e => console.error(e.formattedMessage));
        if (output.errors.some(e => e.severity === 'error')) process.exit(1);
    }

    const contract = output.contracts['CredentialRegistry.sol']['CredentialRegistry'];
    const abi = contract.abi;
    const bytecode = contract.evm.bytecode.object;

    // 连接本地链（Ganache）
    const provider = new ethers.JsonRpcProvider(
        process.env.CHAIN_RPC || 'http://127.0.0.1:8545'
    );

    // 部署者私钥必须通过环境变量注入，禁止把私钥写进源码
    const privateKey = process.env.DEPLOYER_KEY;
    if (!privateKey) {
        console.error('请设置环境变量 DEPLOYER_KEY（部署者私钥），不要硬编码到脚本中');
        process.exit(1);
    }
    const wallet = new ethers.Wallet(privateKey, provider);

    console.log('部署账户:', wallet.address);

    // 部署
    const factory = new ethers.ContractFactory(abi, bytecode, wallet);
    const registry = await factory.deploy();
    await registry.waitForDeployment();

    const address = await registry.getAddress();
    console.log('');
    console.log('========== 部署成功 ==========');
    console.log('合约地址:', address);
    console.log('部署账户:', wallet.address);
    console.log('');
    console.log('请将合约地址填入后端环境变量:');
    console.log(`  export CONTRACT_ADDR=${address}`);
    console.log('');

    // 保存 ABI 供后端使用
    fs.writeFileSync(
        path.join(__dirname, 'CredentialRegistry.json'),
        JSON.stringify({ abi, bytecode, address }, null, 2)
    );
    console.log('ABI 已保存到 CredentialRegistry.json');
}

main().catch(console.error);
