// aesGcmWeb.js
// Usage:
//   const cipherB64 = await encryptString(passphrase, "hello")
//   const plain = await decryptString(passphrase, cipherB64)

function bufToBase64(buf) {
    const bytes = new Uint8Array(buf);
    let binary = "";
    const chunk = 0x8000;
    for (let i = 0; i < bytes.length; i += chunk) {
        binary += String.fromCharCode.apply(null, Array.from(bytes.subarray(i, i + chunk)));
    }
    return btoa(binary);
}

function base64ToBuf(b64) {
    const binary = atob(b64);
    const len = binary.length;
    const bytes = new Uint8Array(len);
    for (let i = 0; i < len; i++) bytes[i] = binary.charCodeAt(i);
    return bytes.buffer;
}

async function deriveKey(passphrase, salt, iterations = 100000, keyLen = 256) {
    const enc = new TextEncoder();
    const passKey = await crypto.subtle.importKey(
        "raw",
        enc.encode(passphrase),
        { name: "PBKDF2" },
        false,
        ["deriveKey"]
    );
    const key = await crypto.subtle.deriveKey(
        {
            name: "PBKDF2",
            salt: salt,
            iterations,
            hash: "SHA-256"
        },
        passKey,
        { name: "AES-GCM", length: keyLen },
        false,
        ["encrypt", "decrypt"]
    );
    return key;
}

export async function encryptString(passphrase, plaintext) {
    const enc = new TextEncoder();
    const salt = crypto.getRandomValues(new Uint8Array(16));
    const iv = crypto.getRandomValues(new Uint8Array(12));
    const key = await deriveKey(passphrase, salt);

    const ct = await crypto.subtle.encrypt(
        {
            name: "AES-GCM",
            iv: iv
        },
        key,
        typeof plaintext === "string" ? enc.encode(plaintext) : plaintext
    );

    // concat salt + iv + ciphertext
    const out = new Uint8Array(salt.byteLength + iv.byteLength + ct.byteLength);
    out.set(salt, 0);
    out.set(iv, salt.byteLength);
    out.set(new Uint8Array(ct), salt.byteLength + iv.byteLength);

    return bufToBase64(out.buffer);
}

export async function decryptString(passphrase, b64) {
    const dec = new TextDecoder();
    const raw = new Uint8Array(base64ToBuf(b64));
    if (raw.length < 16 + 12) throw new Error("ciphertext too short");
    const salt = raw.slice(0, 16);
    const iv = raw.slice(16, 16 + 12);
    const ct = raw.slice(28);

    const key = await deriveKey(passphrase, salt);

    const plainBuf = await crypto.subtle.decrypt(
        {
            name: "AES-GCM",
            iv: iv
        },
        key,
        ct
    );
    return dec.decode(plainBuf);
}
