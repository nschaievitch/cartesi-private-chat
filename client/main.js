import { JsonRpcProvider } from "@ethersproject/providers";
import { ethers } from "ethers";
import { InputBox__factory } from "@cartesi/rollups";
import { runModule } from "./key-gen.js";
import { AES, enc } from "crypto-js";

const HARDHAT_DEFAULT_MNEMONIC =
    "test test test test test test test test test test test junk";

const INPUTBOX_ADDRESS = "0x59b22D57D4f067708AB0c00552767405926dc768";
const DAPP_ADDRESS = "0x70ac08179605AF2D9e75782b8DEcDD3c22aA4D0C";

const HARDHAT_LOCALHOST_RPC_URL = "http://localhost:8545";
const INSPECT_LOCALHOST_URL = "http://localhost:8080/inspect";

const provider = new JsonRpcProvider(HARDHAT_LOCALHOST_RPC_URL);

const getNthSigner = (accountIndex) =>
    ethers.Wallet.fromMnemonic(
        HARDHAT_DEFAULT_MNEMONIC,
        `m/44'/60'/0'/0/${accountIndex}`
    ).connect(provider);

const sendEncodedInput = async (inputBox, jsonLike) => {
    const inputBytes = ethers.utils.toUtf8Bytes(JSON.stringify(jsonLike));
    const tx = await inputBox.addInput(DAPP_ADDRESS, inputBytes);

    console.log(tx);
    const rec = await tx.wait(1);

    console.log(rec);
};

const fetchGroups = async () => {
    const resp = await fetch(`${INSPECT_LOCALHOST_URL}/{"method":"groups"}`);
    const json = await resp.json();

    const payload = json.reports[0].payload;
    return JSON.parse(
        ethers.utils.toUtf8String(ethers.utils.arrayify(payload))
    );
};

const fetchGroupTransitions = async (groupID) => {
    const resp = await fetch(
        `${INSPECT_LOCALHOST_URL}/{"method":"transitions", "id":"${groupID}"}`
    );
    const json = await resp.json();

    const payload = json.reports[0].payload;
    return JSON.parse(
        ethers.utils.toUtf8String(ethers.utils.arrayify(payload))
    );
};

const getInputBox = (signer) =>
    InputBox__factory.connect(INPUTBOX_ADDRESS, signer);

const members = [0, 1, 2].map((i) => getNthSigner(i).address);

const createGroup = async (inputBox, members) => {
    await sendEncodedInput(inputBox, {
        method: "CreateGroup",
        members,
    });
};

const generateDHKeys = async () => {
    const ps = await runModule("key-gen.wasm", [
        "generate",
        `${Math.floor(Math.random() * 10000)}`,
    ]);
    return {
        keys: ps[0],
        r1: ps[1],
    };
};

const computeR2 = async (group, keys) => {
    const ind = group.R1.indexOf(keys.r1);

    console.log(group.R1, keys.r1, ind);

    const r1String = group.R1.join(".");
    const ps = await runModule("key-gen.wasm", [
        "get-r2",
        keys.keys,
        ind,
        r1String,
    ]);

    return ps[0];
};

const computeSecret = async (group, keys) => {
    const ind = group.R1.indexOf(keys.r1);
    const r1String = group.R1.join(".");
    const r2String = group.R2.join(".");
    const ps = await runModule("key-gen.wasm", [
        "get-secret",
        keys.keys,
        ind,
        r1String,
        r2String,
    ]);

    return ps[0];
};

const submitR1 = async (inputBox, groupID, r1) => {
    await sendEncodedInput(inputBox, {
        method: "SubmitR1",
        id: groupID,
        r1Value: r1,
    });
};

const submitR2 = async (inputBox, groupID, r2) => {
    await sendEncodedInput(inputBox, {
        method: "SubmitR2",
        id: groupID,
        r2Value: r2,
    });
};

let signer = getNthSigner(0);
let contract = getInputBox(signer);
let secret, activeGroup;

document.querySelector("#signerNum").onchange = async (e) => {
    if (e.target.value < 0) {
        const provider = new ethers.providers.Web3Provider(
            window.ethereum,
            "any"
        );
        // Prompt user for account connections
        await provider.send("eth_requestAccounts", []);
        signer = provider.getSigner();
    } else {
        signer = getNthSigner(e.target.value);
    }

    document.querySelector("#address").innerText = await signer.getAddress();
    contract = getInputBox(signer);
};

const updateGroups = async () => {
    const table = document.querySelector("#groups");
    const groups = await fetchGroups();
    const address = await signer.getAddress()
    console.log(groups);

    let header = "<tr><td>Group ID</td><td>Members</td><td>Action</td></tr>";
    let rows = Object.entries(groups)
        .filter((g) => g[1].Members.includes(address))
        .map((g) => {
            const ind = g[1].Members.indexOf(address);

            const canJoin = g[1].R1[ind] == null;
            const canSign =
                g[1].R2[ind] == null &&
                g[1].R1.filter((t) => t == null).length == 0;
            const canVisit = g[1].R2.filter((t) => t == null).length == 0;

            let join = `<button class="btn btn-primary" onclick="window.join('${g[0]}')">Join</button>`;
            let sign = `<button class="btn btn-success" onclick="window.sign('${g[0]}')">Sign</button>`;
            let visit = `<button class="btn btn-success" onclick="window.visit('${g[0]}')">Visit</button>`;

            let actions = canJoin ? join : "";
            actions += canSign ? sign : "";
            actions += canVisit ? visit : "";
            actions = actions.length == 0 ? "(Wait for others)" : actions;

            return `<tr> <td> <h6>${g[0]}<h6> </td>  <td>${g[1].Members.join(
                ", "
            )}</td> <td>${actions}</td></tr>`;
        });

    table.innerHTML = header + rows.join("\n");

    if (secret != null) {
        const tr = await fetchGroupTransitions(activeGroup);
        const msgs = document.querySelector("#msgs")
        msgs.innerHTML = tr.map(t => `<strong>${t.Author}</strong>: ${AES.decrypt(t.Action, secret).toString(enc.Utf8)}`).join('<br>')
    }
};

const recoverKeys = (groupID) => {
    const k = localStorage.getItem(groupID);
    const r1 = k.split(".")[1];

    return {
        keys: k,
        r1,
    };
};

window.join = async (groupID) => {
    const k = await generateDHKeys();
    localStorage.setItem(groupID, k.keys);

    await submitR1(contract, groupID, k.r1);
    updateGroups();
};

window.sign = async (groupID) => {
    const keys = recoverKeys(groupID);

    const gs = await fetchGroups();
    const group = gs[groupID];

    const r2 = await computeR2(group, keys);
    submitR2(contract, groupID, r2);

    updateGroups();
};

window.visit = async (groupID) => {
    const keys = recoverKeys(groupID);

    const gs = await fetchGroups();
    const group = gs[groupID];

    const secretComp = await computeSecret(group, keys);
    secret = secretComp;
    activeGroup = groupID;

    const win = document.querySelector("#groupWindow");

    win.innerHTML = `<div id='msgs'> </div> <input class='form-control d-inline' id='txt-send'></input><button class='btn btn-primary' onclick='window.send("${groupID}")'>Send</button>`;
};

window.send = async (groupID) => {
    const inp = document.querySelector("#txt-send")
    await sendEncodedInput(contract, {
        method: "SubmitTransition",
        id: groupID,
        action: AES.encrypt(inp.value, secret).toString()
    });
    inp.value = ""
    console.log("send");
};

document.querySelector("#groupCreate").onclick = async () => {
    const ta = document.querySelector("textarea#addresses");
    const v = ta.value;

    const addresses = v
        .split("\n")
        .filter((t) => t.trim().length > 2)
        .map((t) => t.trim());
    console.log(addresses);
    await createGroup(contract, addresses);
    updateGroups();
};

setInterval(updateGroups, 3000);
