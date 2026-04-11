class ResilverNews extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._refreshInterval = (cfg.refreshIntervalSeconds || 1800) * 1000;
    this._cycleInterval = (cfg.cycleIntervalSeconds || 10) * 1000;
    this._items = [];
    this._index = 0;

    this.className = "flex flex-col justify-center items-center w-full h-full text-gray-300 text-center";
    this.innerHTML = `
      <div class="resilver-news__headline font-light" style="font-size: 3cqmin; transition: opacity 0.6s ease"></div>
    `;

    this._headlineEl = this.querySelector(".resilver-news__headline");

    this._fetch();
    this._fetchTimer = setInterval(() => this._fetch(), this._refreshInterval);
  }

  disconnectedCallback() {
    clearInterval(this._fetchTimer);
    clearInterval(this._cycleTimer);
  }

  async _fetch() {
    try {
      const resp = await fetch("/api/news");
      if (!resp.ok) throw new Error(`HTTP ${resp.status}`);
      this._items = await resp.json();
      this._index = 0;
      this._show();
      this._startCycle();
    } catch (err) {
      this._headlineEl.style.opacity = "1";
      this._headlineEl.textContent = "Unable to load news";
    }
  }

  _startCycle() {
    clearInterval(this._cycleTimer);
    if (this._items.length <= 1) return;
    this._cycleTimer = setInterval(() => this._next(), this._cycleInterval);
  }

  _next() {
    this._headlineEl.style.opacity = "0";
    setTimeout(() => {
      this._index = (this._index + 1) % this._items.length;
      this._show();
    }, 600);
  }

  _show() {
    if (this._items.length === 0) return;
    this._headlineEl.textContent = this._items[this._index].title;
    this._headlineEl.style.opacity = "1";
  }
}

customElements.define("resilver-news", ResilverNews);
