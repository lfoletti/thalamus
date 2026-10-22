"""OpenAI-compatible /v1/embeddings server over a fastembed model, for cortexdb-grpc
(OPENAI_BASE_URL=http://127.0.0.1:<port>/v1). Usage: encoder.py <model> <port>"""
import json, sys
from http.server import BaseHTTPRequestHandler, ThreadingHTTPServer
from fastembed import TextEmbedding

model = TextEmbedding(sys.argv[1])
port = int(sys.argv[2])
print(f"encoder {sys.argv[1]}, port {port}", flush=True)


class Handler(BaseHTTPRequestHandler):
    def do_POST(self):  # noqa: N802
        body = json.loads(self.rfile.read(int(self.headers["Content-Length"])))
        texts = body["input"] if isinstance(body["input"], list) else [body["input"]]
        vecs = [v.tolist() for v in model.embed(texts)]
        out = json.dumps({"object": "list", "model": body.get("model", ""),
                          "data": [{"object": "embedding", "index": i, "embedding": v} for i, v in enumerate(vecs)]}).encode()
        self.send_response(200)
        self.send_header("Content-Type", "application/json")
        self.send_header("Content-Length", str(len(out)))
        self.end_headers()
        self.wfile.write(out)

    def log_message(self, *args):
        pass


ThreadingHTTPServer(("127.0.0.1", port), Handler).serve_forever()
