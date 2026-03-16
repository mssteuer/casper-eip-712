const fs = require("fs");
const path = require("path");

const vectors = {
  generatedAt: new Date().toISOString(),
  note: "Placeholder vector generator scaffold. Replace with ethers.js-based generation in CI or local dev tooling.",
  vectors: []
};

fs.writeFileSync(path.join(__dirname, "..", "tests", "vectors.json"), JSON.stringify(vectors, null, 2));
console.log("wrote tests/vectors.json");
