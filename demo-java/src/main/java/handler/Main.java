package handler;

import org.apache.log4j.Logger;
import org.bouncycastle.asn1.DEROctetString;
import org.hyperledger.fabric.gateway.*;
import org.hyperledger.fabric.gateway.spi.CommitListener;
import org.hyperledger.fabric.gateway.spi.PeerDisconnectEvent;
import org.hyperledger.fabric.sdk.BlockEvent;
import org.hyperledger.fabric.sdk.BlockInfo;


import java.io.IOException;
import java.nio.charset.StandardCharsets;
import java.nio.file.Files;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.security.InvalidKeyException;
import java.security.PrivateKey;
import java.security.cert.CertificateException;
import java.security.cert.X509Certificate;
import java.util.concurrent.TimeoutException;
import java.util.function.Consumer;

/**
 * run this main method and check result
 */
public class Main {
    private static Logger logger = Logger.getLogger(Main.class);

    public static void main(String[] args) throws Exception {
        // Init FabricHelper
        // FabricHelper parses the sdk yaml configuration file and provides the information the user needs
        // when initialize the fabric gateway client
        // Note: Each FabricHelper instance corresponds to one configuration file
        // Please modify the value to the actual path of your sdk yaml configuration file
        FabricHelper helper = new FabricHelper("config/bcs-test-channel-sdk-config.yaml");
        Gateway.Builder builder = initGateway(helper);

        try (Gateway gateway = builder.connect()) {
            // You can modify the channel name instead of using the default, such as "channel"
            Network network = gateway.getNetwork(helper.getDefaultChannel());
            // You can modify the chaincode name instead of using the default, such as "chaincode"
            Contract contract = network.getContract(helper.getDefaultChaincode(helper.getDefaultChannel()));

            // A sample for basic invoke and query methods
            basicTransactionSample(contract);
            // A sample for CommitListener usages
            //commitListenerSample(network, contract);
            // A sample for ContractListener usages
           // contractListenerSample(contract);
            // A sample for BlockListener usages
            //blockListenerSample(network, contract);
        }
    }

    private static void basicTransactionSample(Contract contract) throws ContractException, InterruptedException, TimeoutException {
        logger.info("---basicTransactionSample---");

        // Submit transactions that store state to the ledger.
        String[] data = {"asset1", "1", "62567598498626", "62567598498626", "3604604142873", "107HAXI", "2022-Jul-23T08:43:08 +0000", "10"};
        byte[] invokeResult = contract.submitTransaction("AddCTEwithAsset", data);
        logger.info("insert new data <" + ">" + " success");
        // Evaluate transactions that query state from the ledger.
        byte[] queryResult = contract.evaluateTransaction("ReadAsset", "asset1");
        logger.info("query key <" + "asset1" + "> value is " + new String(queryResult, StandardCharsets.UTF_8));
    }

    private static void commitListenerSample(Network network, Contract contract) throws ContractException, TimeoutException, InterruptedException {
        logger.info("---commitListenerSample---");
        // CommitListener is notified both of commit events received from peers and also of peer communication failures.
        CommitListener listener = new CommitListener() {
            // Called to notify the listener that a given peer has processed a transaction.
            @Override
            public void acceptCommit(BlockEvent.TransactionEvent transactionEvent) {
                logger.info("Commit from peer: " + transactionEvent.getPeer().getName());
                BlockInfo.TransactionEnvelopeInfo.TransactionActionInfo actionInfo = transactionEvent.getTransactionActionInfo(0);
                logger.info("insert new data <" + new String(actionInfo.getChaincodeInputArgs(1)) + ", " + new String(actionInfo.getChaincodeInputArgs(2)) + ">" + " success");
            }

            // Called to notify the listener of a communication failure with a given peer.
            @Override
            public void acceptDisconnect(PeerDisconnectEvent peerDisconnectEvent) {
                logger.info("Fail to connect to url: " + peerDisconnectEvent.getPeer().getUrl());
            }
        };

        Transaction insertTrans = contract.createTransaction("insert");
        network.addCommitListener(listener, network.getChannel().getPeers(), insertTrans.getTransactionId());
        // You can modify the following codes to submit a transaction
        String[] data = {"testuser", "100"};
        insertTrans.submit(data);
        // Removes a previously added transaction commit listener.
        network.removeCommitListener(listener);
    }

    // Contract listener can receive contract events emitted by committed transactions.
    private static void contractListenerSample(Contract contract) throws ContractException, TimeoutException, InterruptedException {
        logger.info("---contractListenerSample---");
        Consumer<ContractEvent> listener = new Consumer<ContractEvent>() {
            @Override
            public void accept(ContractEvent contractEvent) {
                logger.info(String.format("Event name: %s; val: %s", contractEvent.getName(),
                        new String(contractEvent.getPayload().isPresent() ? contractEvent.getPayload().get() : "no value".getBytes())));
            }
        };
        // The listener is only notified of events with name "insert key name"
        //contract.addContractListener(listener, "insert key name");

        // Receive all contract events emitted by committed transactions. Any new contract event won't be received.
        contract.addContractListener(listener);


        for (int i = 0; i < 5; i++) {
            if (i >= 3) {
                contract.removeContractListener(listener);
            }
            // You can modify the following codes to submit a transaction
            String[] data = {"Btestuser" + i, "100" + i};
            byte[] invokeResult = contract.submitTransaction("insert", data);
            logger.info("insert new data <" + data[0] + ", " + data[1] + ">" + " success");

            // You can modify the following codes to evaluate a transaction
            String[] name = {"Btestuser" + i};
            // contract.evaluateTransaction() is recommended for querying. But it won't commit transactions,
            // no events will be emitted.
            byte[] queryResult = contract.submitTransaction("query", name);
            logger.warn("query key <" + name[0] + "> value is " + new String(queryResult, StandardCharsets.UTF_8));
        }

        contract.addContractListener(listener);
        // query a key that doesn't exist
        String[] name = {"Btestuser123"};
        byte[] queryResult = contract.submitTransaction("query", name);
        logger.warn("query key <" + name[0] + "> value is " + new String(queryResult, StandardCharsets.UTF_8));

        contract.removeContractListener(listener);
    }

    // Block listener can receive block events from the network with checkpointing
    private static void blockListenerSample(Network network, Contract contract) throws ContractException, TimeoutException, InterruptedException {
        logger.info("---blockListenerSample---");
        Consumer<BlockEvent> listener = new Consumer<BlockEvent>() {
            @Override
            public void accept(BlockEvent blockEvent) {
                // getBlockNumber() returns the block index number
                // getDataHash() returns data hash value and null if filtered block.
                logger.info(String.format("BlockNumber: %s ; DataHash: %s", blockEvent.getBlockNumber(),
                        new DEROctetString(blockEvent.getDataHash())));
            }
        };

        network.addBlockListener(listener);

        for (int i = 0; i < 5; i++) {
            if (i >= 3) {
                network.removeBlockListener(listener);
            }
            String[] data = {"Ctestuser" + i, "100" + i};
            byte[] invokeResult = contract.submitTransaction("insert", data);
            logger.info("insert new data <" + data[0] + ", " + data[1] + ">" + " success");
        }

        network.removeBlockListener(listener);
    }

    public static Gateway.Builder initGateway(FabricHelper helper) throws IOException, CertificateException, InvalidKeyException {

        // Read from the configuration file by default
        String orgMSP = helper.getOrgMSP();
        // You can modify the cert path instead of reading from the configuration file
        Path certPath = helper.getCertPath();
        // You can modify the private key path instead of reading from the configuration file
        Path privateKeyPath = helper.getPrivateKeyPath();

        // Load cert and private key
        X509Certificate certificate = Identities.readX509Certificate(Files.newBufferedReader(certPath, StandardCharsets.UTF_8));
        PrivateKey privateKey = Identities.readPrivateKey(Files.newBufferedReader(privateKeyPath, StandardCharsets.UTF_8));

        // Init the identity of the gateway client in the wallet
        Wallet wallet = Wallets.newFileSystemWallet(Paths.get("wallet"));
        wallet.put("admin", Identities.newX509Identity(orgMSP, certificate, privateKey));

        Path networkConfigFile = helper.getConfigPath();

        // Init gateway client with identity and config file
        return Gateway.createBuilder()
                // You can modify the directory instead of using wallet/admin.id.
                .identity(wallet, "admin")
                .networkConfig(networkConfigFile);
    }
}