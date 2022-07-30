package handler;

import org.apache.log4j.Logger;
import org.yaml.snakeyaml.Yaml;

import java.io.*;
import java.nio.file.Path;
import java.nio.file.Paths;
import java.util.*;

import javax.json.Json;
import javax.json.JsonObject;

public class FabricHelper {
    private Logger logger = Logger.getLogger(FabricHelper.class);

    private String ymlFileName = "./config.yaml";
    private JsonObject confJson;

    public FabricHelper(String configPath) {
        readYamlConfig(configPath);
    }

    public String getAccessKey() {
        return confJson.getJsonObject("client").getString("organization");
    }

    public String getOrgMSP() {
        String prefix = "MSP";
        return getAccessKey() + prefix;
    }

    public String getDefaultChannel() {
        JsonObject channels = confJson.getJsonObject("channels");
        // Get the only channel from config
        String chanNameTemp = channels.keySet().toString();
        return chanNameTemp.substring(chanNameTemp.indexOf("[") + 1, chanNameTemp.indexOf("]"));
    }

    public String getDefaultChaincode(String channelName) {
        JsonObject channels = confJson.getJsonObject("channels");

        // Get the only chaincode name from config
        String codeNameTemp = channels.getJsonObject(channelName).getJsonArray("chaincodes").getString(0);
        return codeNameTemp.substring(0, codeNameTemp.indexOf(":"));
    }

    public Path getCertPath() {
        String msp = getCryptoPath(this.ymlFileName);
        return Paths.get(getFirstFileFromDir(msp, "signcerts"));
    }

    public Path getPrivateKeyPath() {
        String msp = getCryptoPath(this.ymlFileName);
        return Paths.get(getFirstFileFromDir(msp, "keystore"));
    }

    public Path getConfigPath() {
        return Paths.get(this.ymlFileName);
    }

    public void readYamlConfig(String configPath) {
        if (configPath == null || configPath.equals("")) {
            logger.error("config path is empty! please input correct ymal path!");
        }
        this.ymlFileName = configPath;

        try {
            InputStream stream = new FileInputStream(new File(this.ymlFileName));
            Yaml yaml = new Yaml();
            Map<String, Object> confYaml = yaml.load(stream);
            confJson = Json.createObjectBuilder(confYaml).build();
        } catch (Exception e) {
            e.printStackTrace();
            logger.error(e.getMessage());
        }
    }


    private String getCryptoPath(String configFile) {
        JsonObject root = null;
        String orgId = getAccessKey();
        try {
            InputStream stream = new FileInputStream(configFile);
            Yaml yaml = new Yaml();
            Map<String, Object> map = yaml.load(stream);
            root = Json.createObjectBuilder(map).build();
        } catch (FileNotFoundException e) {
            // TODO Auto-generated catch block
            e.printStackTrace();
        }

        JsonObject orgs = root.getJsonObject("organizations");
        for (Map.Entry o : orgs.entrySet()) {
            if (orgId.equals(o.getKey())) {
                JsonObject v = (JsonObject) o.getValue();
                return v.getString("cryptoPath");
            }
        }
        return "";
    }

    private String getFirstFileFromDir(String path, String sub) {
        File dir = new File(path + "/" + sub);
        if (!dir.exists()) {
            logger.error("directory is not exist. path:" + dir);
            return "";
        }

        for (File f : dir.listFiles()) {
            return f.getAbsolutePath();
        }
        logger.error("cannot get any file from directory. path:" + dir);
        return "";
    }

}