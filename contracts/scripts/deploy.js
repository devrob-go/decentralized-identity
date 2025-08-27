const { ethers } = require("hardhat");

async function main() {
    console.log("Deploying DID Registry contract...");

    // Get the deployer account
    const [deployer] = await ethers.getSigners();
    console.log("Deploying contracts with account:", deployer.address);
    console.log("Account balance:", (await deployer.getBalance()).toString());

    // Deploy the DID Registry contract
    const DIDRegistry = await ethers.getContractFactory("DIDRegistry");
    const didRegistry = await DIDRegistry.deploy();
    await didRegistry.deployed();

    console.log("DID Registry deployed to:", didRegistry.address);

    // Add the deployer as an authorized operator
    console.log("Adding deployer as authorized operator...");
    const addOperatorTx = await didRegistry.addAuthorizedOperator(deployer.address);
    await addOperatorTx.wait();
    console.log("Deployer added as authorized operator");

    // Verify the deployment
    console.log("Verifying deployment...");
    const stats = await didRegistry.getStats();
    console.log("Contract stats:", {
        total: stats.total.toString(),
        active: stats.active.toString(),
        revoked: stats.revoked.toString(),
    });

    // Save deployment info
    const deploymentInfo = {
        network: hre.network.name,
        contract: "DIDRegistry",
        address: didRegistry.address,
        deployer: deployer.address,
        timestamp: new Date().toISOString(),
        blockNumber: await deployer.provider.getBlockNumber(),
    };

    console.log("Deployment successful!");
    console.log("Contract address:", didRegistry.address);
    console.log("Network:", hre.network.name);
    console.log("Block number:", deploymentInfo.blockNumber);

    // Export deployment info for other scripts
    if (hre.network.name !== "hardhat") {
        console.log("\nDeployment info saved to deployment-info.json");
        const fs = require("fs");
        fs.writeFileSync(
            "deployment-info.json",
            JSON.stringify(deploymentInfo, null, 2)
        );
    }

    return didRegistry;
}

// Execute deployment
main()
    .then(() => process.exit(0))
    .catch((error) => {
        console.error("Deployment failed:", error);
        process.exit(1);
    });
