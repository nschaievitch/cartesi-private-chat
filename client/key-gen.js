import { WASI } from "@runno/wasi";

export const runModule = async (moduleURL, args) => {
    let resp = [];

    console.log(args)
    const result = WASI.start(fetch(moduleURL), {
        args: ["main.wasm", ...args],
        stdout: (out) => resp.push(out.trim()),
        stderr: (err) => console.error("stderr", err),
        stdin: () => prompt("stdin:"),
        fs: {},
    });

    await result;

    return resp;
};