const form = document.querySelector("#check-form");
const urlsInput = document.querySelector("#urls");
const workersInput = document.querySelector("#workers");
const timeoutInput = document.querySelector("#timeout_ms");
const retriesInput = document.querySelector("#retries");
const message = document.querySelector("#message");
const requestState = document.querySelector("#request-state");
const summary = document.querySelector("#summary");
const resultsBody = document.querySelector("#results-body");
const submitButton = document.querySelector("#submit");
const fillDemoButton = document.querySelector("#fill-demo");

fillDemoButton.addEventListener("click", () => {
  urlsInput.value = [
    "https://example.com",
    "golang.org",
    "https://httpbin.org/status/404",
    "https://httpbin.org/delay/2",
  ].join("\n");
});

form.addEventListener("submit", async (event) => {
  event.preventDefault();

  const urls = urlsInput.value
    .split("\n")
    .map((value) => value.trim())
    .filter(Boolean);

  if (urls.length === 0) {
    setState("error", "Add at least one URL.");
    renderSummary({ ok: 0, fail: 0, total: 0 });
    renderPlaceholder("Run a batch to populate results.");
    return;
  }

  const payload = {
    urls,
    workers: parseInteger(workersInput.value),
    timeout_ms: parseInteger(timeoutInput.value),
    retries: parseInteger(retriesInput.value),
  };

  submitButton.disabled = true;
  setState("loading", "Checking URLs...");
  renderPlaceholder("Waiting for API response...");

  try {
    const response = await fetch("/api/check", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(payload),
    });

    const isJSON = response.headers.get("content-type")?.includes("application/json");
    const data = isJSON ? await response.json() : null;

    if (!response.ok) {
      const text = data ? JSON.stringify(data) : await response.text();
      throw new Error(text || "Request failed");
    }

    renderSummary(data);
    renderResults(data.results || []);
    setState("success", `Completed ${data.total} checks.`);
  } catch (error) {
    renderSummary({ ok: 0, fail: 0, total: 0 });
    renderPlaceholder("Request failed. Check the server and try again.");
    setState("error", error.message || "Request failed");
  } finally {
    submitButton.disabled = false;
  }
});

function parseInteger(value) {
  const parsed = Number.parseInt(value, 10);
  return Number.isFinite(parsed) ? parsed : 0;
}

function setState(kind, text) {
  requestState.className = `request-state ${kind}`;
  requestState.textContent = kind;
  message.textContent = text;
}

function renderSummary(data) {
  summary.classList.toggle("empty", data.total === 0);
  summary.innerHTML = `
    <article>
      <span>OK</span>
      <strong>${data.ok ?? 0}</strong>
    </article>
    <article>
      <span>Fail</span>
      <strong>${data.fail ?? 0}</strong>
    </article>
    <article>
      <span>Total</span>
      <strong>${data.total ?? 0}</strong>
    </article>
  `;
}

function renderResults(results) {
  if (!results.length) {
    renderPlaceholder("No results returned.");
    return;
  }

  resultsBody.innerHTML = results
    .map((item) => {
      const outcomeClass = item.ok ? "ok" : "fail";
      const outcomeText = item.ok ? "OK" : item.error || "FAIL";
      const status = item.status ? item.status : "n/a";
      return `
        <tr>
          <td class="url-cell">${escapeHTML(item.url)}</td>
          <td>${status}</td>
          <td>${item.attempts}</td>
          <td>${item.duration_ms} ms</td>
          <td><span class="outcome ${outcomeClass}">${escapeHTML(outcomeText)}</span></td>
        </tr>
      `;
    })
    .join("");
}

function renderPlaceholder(text) {
  resultsBody.innerHTML = `
    <tr class="placeholder">
      <td colspan="5">${escapeHTML(text)}</td>
    </tr>
  `;
}

function escapeHTML(value) {
  return String(value)
    .replaceAll("&", "&amp;")
    .replaceAll("<", "&lt;")
    .replaceAll(">", "&gt;")
    .replaceAll('"', "&quot;")
    .replaceAll("'", "&#39;");
}
