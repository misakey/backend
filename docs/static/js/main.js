var h2 = document.querySelectorAll("h2")
for (var i = 0; i < h2.length; i++) {
    // add come to TOC arrow to all headers excepted the TOC header itself
    if (h2[i].textContent.toLowerCase() != "Table of Contents".toLowerCase()) {
        h2[i].innerHTML = `<a class="toc_a" href="#toc">⬆</a>${h2[i].innerHTML}`;
    }
}

var h = document.querySelectorAll("h3, h4")
for (var i = 0; i < h.length; i++) {
    // add come to TOC arrow to all headers excepted the TOC header itself
    if (h[i].textContent.toLowerCase().endsWith('response')) {
        h[i].innerHTML = `<a class="toc_a" href = "#${h[i].id}">📥</a> ${h[i].innerHTML} `;
    } else if (h[i].textContent.toLowerCase().endsWith('request')) {
        h[i].innerHTML = `<a class="toc_a" href = "#${h[i].id}">📤</a> ${h[i].innerHTML} `;
    } else if (h[i].textContent.toLowerCase().startsWith('method name')) {
        h[i].innerHTML = `<a class="toc_a" href = "#${h[i].id}">🔐</a> ${h[i].innerHTML} `;
    } else if (h[i].textContent.toLowerCase().includes('notable')) {
        h[i].innerHTML = `<a class="toc_a" href = "#${h[i].id}">❌</a> ${h[i].innerHTML} `;
    }
}