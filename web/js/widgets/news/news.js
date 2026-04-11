class ResilverNews extends HTMLElement {
  connectedCallback() {
    const cfg = JSON.parse(this.dataset.config || "{}");
    this._refreshInterval = (cfg.refreshIntervalSeconds || 1800) * 1000;
    this._cycleInterval = (cfg.cycleIntervalSeconds || 10) * 1000;
    this._items = [];
    this._index = 0;

    this.className =
      "flex justify-center items-center w-full h-full text-gray-300";
    this.innerHTML = `
      <div class="resilver-news__content flex items-center rounded-lg overflow-hidden" style="transition: opacity 0.6s ease; max-width: 60%">
        <img class="resilver-news__image hidden flex-shrink-0 object-cover rounded-l-lg" style="width: 40cqmin; height: 40cqmin" />
        <div class="flex flex-col justify-center px-[2cqmin] flex-1 min-w-0 pl-3 gap-1">
          <div class="resilver-news__headline font-light leading-snug" style="font-size: 2.5cqw"></div>
          <div class="resilver-news__source opacity-40" style="font-size: 1.5cqw"></div>
        </div>
      </div>
    `;

    this._contentEl = this.querySelector(".resilver-news__content");
    this._imageEl = this.querySelector(".resilver-news__image");
    this._headlineEl = this.querySelector(".resilver-news__headline");
    this._sourceEl = this.querySelector(".resilver-news__source");

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
      this._contentEl.style.opacity = "1";
      this._headlineEl.textContent = "Unable to load news";
      this._sourceEl.textContent = "";
      this._imageEl.classList.add("hidden");
    }
  }

  _startCycle() {
    clearInterval(this._cycleTimer);
    if (this._items.length <= 1) return;
    this._cycleTimer = setInterval(() => this._next(), this._cycleInterval);
  }

  _next() {
    this._contentEl.style.opacity = "0";
    setTimeout(() => {
      this._index = (this._index + 1) % this._items.length;
      this._show();
    }, 600);
  }

  _show() {
    if (this._items.length === 0) return;
    const item = this._items[this._index];
    this._headlineEl.textContent = item.title;
    this._sourceEl.textContent = item.source || "";

    if (item.image) {
      this._imageEl.src = item.image;
      this._imageEl.classList.remove("hidden");
    } else {
      this._imageEl.classList.add("hidden");
      this._imageEl.removeAttribute("src");
    }

    this._contentEl.style.opacity = "1";
  }
}

customElements.define("resilver-news", ResilverNews);
