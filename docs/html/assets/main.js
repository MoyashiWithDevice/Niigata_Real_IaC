/*
 * IACForge Reference - main script
 * Handles: sidebar, mobile nav, search, TOC generation, scrollspy,
 *          YAML syntax highlighting, copy buttons.
 * Dependency-free.
 */
(function () {
  "use strict";

  var ICONS = {
    chevron:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2.5" stroke-linecap="round" stroke-linejoin="round"><polyline points="6 9 12 15 18 9"></polyline></svg>',
    search:
      '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><circle cx="11" cy="11" r="8"></circle><line x1="21" y1="21" x2="16.65" y2="16.65"></line></svg>',
    menu: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round"><line x1="3" y1="6" x2="21" y2="6"></line><line x1="3" y1="12" x2="21" y2="12"></line><line x1="3" y1="18" x2="21" y2="18"></line></svg>',
    doc: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M14 2H6a2 2 0 0 0-2 2v16a2 2 0 0 0 2 2h12a2 2 0 0 0 2-2V8z"></path><polyline points="14 2 14 8 20 8"></polyline></svg>',
    note: '<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round"><path d="M18 3a3 3 0 0 0-3 3v12a3 3 0 0 0 3 3 3 3 0 0 0 3-3 3 3 0 0 0-3-3H6a3 3 0 0 0-3 3 3 3 0 0 0 3 3 3 3 0 0 0 3-3V6a3 3 0 0 0-3-3 3 3 0 0 0-3 3 3 3 0 0 0 3 3h12a3 3 0 0 0 3-3 3 3 0 0 0-3-3z"></path></svg>',
  };

  /* ------------------------------------------------------------------ *
   * Top bar chrome: icons + hamburger
   * ------------------------------------------------------------------ */
  function initChrome() {
    document.querySelectorAll(".topbar-search .search-icon").forEach(function (el) {
      el.innerHTML = ICONS.search;
    });
    document.querySelectorAll(".menu-toggle").forEach(function (el) {
      el.innerHTML = ICONS.menu;
    });
    document.querySelectorAll(".topbar-brand .topbar-logo").forEach(function (el) {
      if (!el.textContent.trim()) el.textContent = "I";
    });
    document.querySelectorAll(".sidebar-links .link-icon").forEach(function (el) {
      el.innerHTML = ICONS.doc;
    });
  }

  /* ------------------------------------------------------------------ *
   * Sidebar: collapse sections + mobile toggle
   * ------------------------------------------------------------------ */
  function initSidebar() {
    document.querySelectorAll(".sidebar-section").forEach(function (sec) {
      var title = sec.querySelector(".sidebar-section-title");
      if (!title) return;
      var chev = document.createElement("span");
      chev.className = "chevron";
      chev.innerHTML = ICONS.chevron;
      title.appendChild(chev);
      title.addEventListener("click", function () {
        sec.classList.toggle("collapsed");
        try {
          localStorage.setItem("iacforge-sidebar-" + title.textContent.trim(), sec.classList.contains("collapsed") ? "1" : "0");
        } catch (e) {}
      });
      // Restore state
      var key = "iacforge-sidebar-" + title.textContent.trim();
      try {
        if (localStorage.getItem(key) === "1") sec.classList.add("collapsed");
      } catch (e) {}
    });

    // If a link inside a collapsed section is active, expand it.
    document.querySelectorAll(".sidebar-section").forEach(function (sec) {
      if (sec.querySelector("a.active")) sec.classList.remove("collapsed");
    });

    var toggle = document.querySelector(".menu-toggle");
    var sidebar = document.querySelector(".sidebar");
    if (toggle && sidebar) {
      toggle.addEventListener("click", function () {
        sidebar.classList.toggle("open");
      });
      document.querySelectorAll(".sidebar a").forEach(function (a) {
        a.addEventListener("click", function () {
          if (window.innerWidth <= 860) sidebar.classList.remove("open");
        });
      });
    }
  }

  /* ------------------------------------------------------------------ *
   * Search
   * ------------------------------------------------------------------ */
  function initSearch() {
    var input = document.getElementById("search-input");
    if (!input) return;
    var dropdown = document.getElementById("search-dropdown");

    function closeDropdown() {
      dropdown.classList.remove("open");
    }

    function openDropdown() {
      dropdown.classList.add("open");
    }

    input.addEventListener("input", function () {
      var q = input.value.trim().toLowerCase();
      if (!q) {
        closeDropdown();
        clearHighlights();
        return;
      }
      var idx = window.IACFORGE_SEARCH_INDEX || [];
      var results = idx
        .map(function (item) {
          var score = 0;
          var hay = (item.title + " " + item.section + " " + item.description + " " + item.keywords).toLowerCase();
          if (hay.indexOf(q) !== -1) {
            score = hay.indexOf(q) === 0 ? 3 : 2;
            if (item.title.toLowerCase().indexOf(q) !== -1) score += 2;
          }
          if (item.title.toLowerCase().indexOf(q) !== -1) score += 1;
          return { item: item, score: score };
        })
        .filter(function (r) {
          return r.score > 0;
        })
        .sort(function (a, b) {
          return b.score - a.score;
        });

      dropdown.innerHTML = "";
      if (results.length === 0) {
        var empty = document.createElement("div");
        empty.className = "search-dropdown-empty";
        empty.textContent = "No results found for \u201c" + input.value + "\u201d";
        dropdown.appendChild(empty);
        openDropdown();
        return;
      }
      results.slice(0, 8).forEach(function (r) {
        var a = document.createElement("a");
        a.className = "search-dropdown-item";
        a.href = r.item.path;
        var t = document.createElement("div");
        t.className = "sdi-title";
        t.textContent = r.item.title;
        var s = document.createElement("div");
        s.className = "sdi-desc";
        s.textContent = r.item.section + " \u00b7 " + r.item.description;
        a.appendChild(t);
        a.appendChild(s);
        dropdown.appendChild(a);
      });
      openDropdown();
    });

    document.addEventListener("click", function (e) {
      if (!dropdown.contains(e.target) && e.target !== input) closeDropdown();
    });

    document.addEventListener("keydown", function (e) {
      if (e.key === "Escape") {
        closeDropdown();
        input.blur();
      }
    });
  }

  /* ------------------------------------------------------------------ *
   * In-page highlight of search hits
   * ------------------------------------------------------------------ */
  function clearHighlights() {
    document.querySelectorAll(".main-content mark.search-hit").forEach(function (m) {
      var parent = m.parentNode;
      parent.replaceChild(document.createTextNode(m.textContent), m);
      parent.normalize();
    });
  }

  /* ------------------------------------------------------------------ *
   * "On this page" TOC + scrollspy
   * ------------------------------------------------------------------ */
  function initTOC() {
    var content = document.querySelector(".main-content");
    var tocEl = document.getElementById("page-toc");
    if (!content || !tocEl) return;

    var headings = content.querySelectorAll("h2, h3");
    if (headings.length === 0) return;

    var ul = document.createElement("ul");
    ul.className = "toc-list";

    headings.forEach(function (h) {
      var id = h.id || slugify(h.textContent);
      h.id = id;
      var li = document.createElement("li");
      var a = document.createElement("a");
      a.href = "#" + id;
      a.textContent = h.textContent.replace(/[\u00a0\s]+$/g, "");
      if (h.tagName === "H3") a.className = "toc-h3";
      li.appendChild(a);
      ul.appendChild(li);
    });

    tocEl.appendChild(ul);

    // Scrollspy
    var links = Array.prototype.slice.call(ul.querySelectorAll("a"));
    var inView = function (el) {
      var r = el.getBoundingClientRect();
      return r.top <= 90;
    };

    var onScroll = function () {
      var current = null;
      headings.forEach(function (h) {
        if (inView(h)) current = h;
      });
      // If near bottom, activate last heading.
      if (window.innerHeight + window.scrollY >= document.body.scrollHeight - 80 && headings.length > 0) {
        current = headings[headings.length - 1];
      }
      links.forEach(function (a) {
        a.classList.toggle("active", !!current && a.getAttribute("href") === "#" + current.id);
      });
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    onScroll();
  }

  function slugify(text) {
    return text
      .toLowerCase()
      .trim()
      .replace(/[^\w\s-]/g, "")
      .replace(/\s+/g, "-")
      .replace(/-+/g, "-");
  }

  /* ------------------------------------------------------------------ *
   * Anchor links on headings
   * ------------------------------------------------------------------ */
  function initAnchors() {
    document.querySelectorAll(".main-content h1, .main-content h2, .main-content h3, .main-content h4").forEach(function (h) {
      if (!h.id) h.id = slugify(h.textContent);
      var a = document.createElement("a");
      a.className = "anchor";
      a.href = "#" + h.id;
      a.setAttribute("aria-label", "Link to this section");
      a.textContent = "#";
      h.appendChild(a);
    });
  }

  /* ------------------------------------------------------------------ *
   * YAML syntax highlighting
   * ------------------------------------------------------------------ */
  function highlightCode() {
    document.querySelectorAll("pre.language-yaml code, pre code.language-yaml").forEach(function (code) {
      code.innerHTML = highlightYAML(code.textContent);
    });
  }

  function esc(html) {
    return html.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;");
  }

  function span(cls, text) {
    return '<span class="' + cls + '">' + text + "</span>";
  }

  function highlightYAML(src) {
    var lines = src.split("\n");
    var out = [];
    for (var i = 0; i < lines.length; i++) {
      var line = lines[i];
      // Comments
      var hashIdx = line.indexOf("#");
      var comment = "";
      var body = line;
      if (hashIdx !== -1) {
        // Don't treat "#" inside quotes as a comment.
        var inS = false;
        var inD = false;
        var realHash = -1;
        for (var c = 0; c < line.length; c++) {
          var ch = line[c];
          if (ch === "'" && !inD) inS = !inS;
          else if (ch === '"' && !inS) inD = !inD;
          else if (ch === "#" && !inS && !inD) {
            realHash = c;
            break;
          }
        }
        if (realHash !== -1) {
          body = line.slice(0, realHash);
          comment = span("tok-comment", esc(line.slice(realHash)));
        }
      }

      var result = highlightLine(body) + comment;
      out.push(result);
    }
    return out.join("\n");
  }

  function highlightLine(line) {
    // Highlight the key part (word before ': ') if present.
    var m = /^(\s*)(-?\s*)(["']?)([A-Za-z0-9_.\/-]+)\3(\s*:\s*)(.*)$/.exec(line);
    if (m) {
      var indent = m[1];
      var dash = m[2];
      var quote = m[3];
      var key = m[4];
      var colon = m[5];
      var rest = m[6];
      var keyHtml =
        span("tok-punct", esc(indent + dash)) +
        (quote ? span("tok-string", esc(quote + key + quote)) : span("tok-key", esc(key))) +
        span("tok-punct", esc(colon));
      return keyHtml + highlightValue(rest);
    }
    // Sequence item with no key (just "- value")
    var seq = /^(\s*-)(\s*)(.*)$/.exec(line);
    if (seq) {
      return span("tok-punct", esc(seq[1])) + highlightValue(seq[3]);
    }
    // Standalone key-less values, document markers, directives
    return highlightValue(line);
  }

  function highlightValue(v) {
    if (v === "") return "";
    var trimmed = v;
    var out = "";
    var i = 0;
    while (i < trimmed.length) {
      var ch = trimmed[i];
      if (ch === " " || ch === "\t") {
        out += esc(ch);
        i++;
        continue;
      }
      if (ch === "#") {
        out += span("tok-comment", esc(trimmed.slice(i)));
        break;
      }
      if (ch === "'" || ch === '"') {
        var quote = ch;
        var end = trimmed.indexOf(quote, i + 1);
        if (end === -1) end = trimmed.length;
        out += span("tok-string", esc(trimmed.slice(i, end + 1)));
        i = end + 1;
        continue;
      }
      // Number or boolean
      var numMatch = /^-?\d+(\.\d+)?$/.exec(trimmed.slice(i));
      if (numMatch) {
        out += span("tok-number", esc(numMatch[0]));
        i += numMatch[0].length;
        continue;
      }
      var boolMatch = /^(true|false|null|yes|no|on|off)\b/i.exec(trimmed.slice(i));
      if (boolMatch) {
        out += span("tok-bool", esc(boolMatch[1]));
        i += boolMatch[1].length;
        continue;
      }
      // Bare word or punctuation: consume until whitespace / comment
      var wordEnd = i;
      while (wordEnd < trimmed.length && trimmed[wordEnd] !== " " && trimmed[wordEnd] !== "\t" && trimmed[wordEnd] !== "#") {
        wordEnd++;
      }
      out += esc(trimmed.slice(i, wordEnd));
      i = wordEnd;
    }
    return out;
  }

  /* ------------------------------------------------------------------ *
   * Copy buttons on code blocks
   * ------------------------------------------------------------------ */
  function initCopyButtons() {
    document.querySelectorAll("pre code").forEach(function (code) {
      var pre = code.parentNode;
      if (pre.querySelector(".copy-btn")) return;
      var btn = document.createElement("button");
      btn.className = "copy-btn";
      btn.textContent = "Copy";
      pre.appendChild(btn);
      btn.addEventListener("click", function () {
        var text = code.textContent;
        var done = function () {
          btn.textContent = "Copied";
          btn.classList.add("copied");
          setTimeout(function () {
            btn.textContent = "Copy";
            btn.classList.remove("copied");
          }, 1600);
        };
        if (navigator.clipboard && navigator.clipboard.writeText) {
          navigator.clipboard.writeText(text).then(done).catch(function () {
            fallbackCopy(text);
            done();
          });
        } else {
          fallbackCopy(text);
          done();
        }
      });
    });
  }

  function fallbackCopy(text) {
    var ta = document.createElement("textarea");
    ta.value = text;
    ta.style.position = "fixed";
    ta.style.opacity = "0";
    document.body.appendChild(ta);
    ta.select();
    try {
      document.execCommand("copy");
    } catch (e) {}
    document.body.removeChild(ta);
  }

  /* ------------------------------------------------------------------ *
   * Init
   * ------------------------------------------------------------------ */
  document.addEventListener("DOMContentLoaded", function () {
    initChrome();
    initSidebar();
    initSearch();
    initTOC();
    initAnchors();
    highlightCode();
    initCopyButtons();
  });
})();
