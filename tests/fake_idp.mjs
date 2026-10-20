// A fake OpenID Connect provider for the benches: discovery, JWKS, an
// authorize endpoint that signs the one employee in without a question
// (code flow with PKCE, or implicit id_token for the extension), and a
// token endpoint that checks the PKCE verifier. RS256 with a key made at
// start. Usage: node fake_idp.mjs --port 4xxxx [--client-id thalamus-cli]
// [--audience thalamus] [--email lfoletti@acme.example]
import { createServer } from "node:http";
import { generateKeyPairSync, sign, randomBytes, createHash } from "node:crypto";

const arg = (name, def) => { const i = process.argv.indexOf("--" + name); return i > 0 ? process.argv[i + 1] : def; };
const port = Number(arg("port", "0"));
const clientID = arg("client-id", "thalamus-cli");
const audience = arg("audience", "thalamus");
const email = arg("email", "lfoletti@acme.example");
const { privateKey, publicKey } = generateKeyPairSync("rsa", { modulusLength: 2048 });
const jwk = publicKey.export({ format: "jwk" });
const b64 = (s) => Buffer.from(s).toString("base64url");
const codes = new Map();
let issuer = "";

const token = (nonce) => {
  const now = Math.floor(Date.now() / 1000);
  const head = b64(JSON.stringify({ alg: "RS256", kid: "k1", typ: "JWT" }));
  const body = b64(JSON.stringify({ iss: issuer, aud: audience, iat: now, exp: now + 3600, nonce, email }));
  return head + "." + body + "." + sign("sha256", Buffer.from(head + "." + body), privateKey).toString("base64url");
};

const server = createServer((req, res) => {
  const url = new URL(req.url, issuer);
  const json = (v) => { res.setHeader("Content-Type", "application/json"); res.end(JSON.stringify(v)); };
  switch (url.pathname) {
    case "/.well-known/openid-configuration":
      return json({ issuer, authorization_endpoint: issuer + "/authorize", token_endpoint: issuer + "/token", jwks_uri: issuer + "/jwks" });
    case "/jwks":
      return json({ keys: [{ kty: "RSA", kid: "k1", alg: "RS256", use: "sig", n: jwk.n, e: jwk.e }] });
    case "/authorize": {
      const q = url.searchParams;
      const back = new URL(q.get("redirect_uri"));
      if (q.get("client_id") !== clientID) {
        back.searchParams.set("error", "unauthorized_client");
        back.searchParams.set("state", q.get("state") || "");
      } else if (q.get("response_type") === "code") {
        const code = randomBytes(8).toString("hex");
        codes.set(code, { nonce: q.get("nonce"), challenge: q.get("code_challenge") });
        back.searchParams.set("code", code);
        back.searchParams.set("state", q.get("state") || "");
      } else {
        back.hash = new URLSearchParams({ id_token: token(q.get("nonce") || ""), state: q.get("state") || "" }).toString();
      }
      res.statusCode = 302;
      res.setHeader("Location", back.toString());
      return res.end();
    }
    case "/token": {
      let body = "";
      req.on("data", (c) => (body += c));
      req.on("end", () => {
        const f = new URLSearchParams(body);
        const c = codes.get(f.get("code"));
        const challenge = createHash("sha256").update(f.get("code_verifier") || "").digest("base64url");
        if (!c || c.challenge !== challenge || f.get("grant_type") !== "authorization_code" || f.get("client_id") !== clientID) {
          res.statusCode = 400;
          return json({ error: "invalid_grant" });
        }
        codes.delete(f.get("code"));
        json({ id_token: token(c.nonce), token_type: "Bearer" });
      });
      return;
    }
    default:
      res.statusCode = 404;
      res.end("not found");
  }
});
server.listen(port, "127.0.0.1", () => {
  issuer = "http://127.0.0.1:" + server.address().port;
  console.log("fake idp at " + issuer);
});
