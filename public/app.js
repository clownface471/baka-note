// ======================================================
// 1. CONFIG
// ======================================================
const API = { journal: "/api/journal", bucket: "/api/bucket", memories: "/api/memories", deep: "/api/deep" };
const START_DATE = new Date("2024-01-01"); // GANTI TANGGAL JADIAN DI SINI!
const nameInput = document.getElementById('myName');

// ======================================================
// 2. THEME & IDENTITY
// ======================================================

// --- Theme Logic ---
const btnTheme = document.getElementById('btnTheme');
const savedTheme = localStorage.getItem('theme');
if (savedTheme === 'dark') {
    document.documentElement.setAttribute('data-theme', 'dark');
    if(btnTheme) btnTheme.textContent = '☀️';
} else {
    if(btnTheme) btnTheme.textContent = '🌙';
}

function toggleTheme() {
    const current = document.documentElement.getAttribute('data-theme');
    if (current === 'dark') {
        document.documentElement.setAttribute('data-theme', 'light');
        localStorage.setItem('theme', 'light');
        btnTheme.textContent = '🌙';
    } else {
        document.documentElement.setAttribute('data-theme', 'dark');
        localStorage.setItem('theme', 'dark');
        btnTheme.textContent = '☀️';
    }
}

// --- Identity Logic ---
function formatName(str) { return str ? str.trim().charAt(0).toUpperCase() + str.trim().slice(1).toLowerCase() : ""; }
if(localStorage.getItem('my_username')) nameInput.value = localStorage.getItem('my_username');
nameInput.addEventListener('change', () => { 
    nameInput.value = formatName(nameInput.value);
    localStorage.setItem('my_username', nameInput.value); 
    // Refresh current tab
    const activeTab = document.querySelector('.nav-btn.active');
    if(activeTab) {
        if(activeTab.getAttribute('onclick').includes('journal')) loadJournal();
        if(activeTab.getAttribute('onclick').includes('bucket')) loadBucket();
    }
});

// --- Clocks & Counter ---
function updateClocks() {
    const now = new Date();
    // 1. Local
    document.getElementById('clock-local').textContent = "YOU " + now.toLocaleTimeString('en-US', {hour:'2-digit',minute:'2-digit',hour12:false});
    
    // 2. Adaptive (SG/ID)
    const offset = now.getTimezoneOffset(); // -420 (ID), -480 (SG)
    let targetZone = 'Asia/Singapore'; let targetLabel = 'SG';
    if (offset === -480) { targetZone = 'Asia/Jakarta'; targetLabel = 'ID'; }
    
    const partnerTime = now.toLocaleTimeString('en-US', {timeZone: targetZone, hour:'2-digit', minute:'2-digit', hour12:false});
    document.getElementById('clock-sg').textContent = targetLabel + " " + partnerTime;

    // 3. Counter
    const diff = now - START_DATE;
    const days = Math.floor(diff / (1000 * 60 * 60 * 24));
    document.getElementById('days-counter').textContent = `${days} DAYS TRACKING`;
}
setInterval(updateClocks, 1000); updateClocks();

// ======================================================
// 3. NAVIGATION
// ======================================================
function switchTab(id) {
    document.querySelectorAll('.view-section').forEach(el => { el.classList.add('hidden'); el.classList.remove('active'); });
    const target = document.getElementById(`view-${id}`);
    if(target) { target.classList.remove('hidden'); target.classList.add('active'); }
    
    document.querySelectorAll('.nav-btn').forEach(el => el.classList.remove('active'));
    const btn = document.querySelector(`.nav-btn[onclick="switchTab('${id}')"]`);
    if(btn) btn.classList.add('active');
    
    if(id==='journal') loadJournal();
    if(id==='bucket') loadBucket();
    if(id==='memories') loadMemories();
    if(id==='deep') loadDeep();
}

// ======================================================
// 4. JOURNAL (CLEAN HTML)
// ======================================================
async function loadJournal() {
    const feed = document.getElementById('feed');
    const me = (nameInput.value || "").toLowerCase();
    const wasAtBottom = feed.scrollHeight - feed.scrollTop <= feed.clientHeight + 50;

    try {
        const res = await fetch(API.journal); const data = await res.json();
        const chatData = data.reverse();
        let html = ""; let lastDate = "";

        chatData.forEach(e => {
            const dateObj = new Date(e.created_at);
            const dateStr = dateObj.toLocaleDateString('en-US', { weekday: 'short', day: '2-digit', month: 'short' });
            const timeStr = dateObj.toLocaleTimeString('en-US', {hour:'2-digit', minute:'2-digit', hour12:false});

            if (dateStr !== lastDate) {
                html += `<div class="date-separator"><span>${dateStr}</span></div>`;
                lastDate = dateStr;
            }

            const isMe = e.author.toLowerCase() === me;
            const alignClass = isMe ? 'msg-right' : 'msg-left';
            const displayName = isMe ? 'You' : e.author;
            
            html += `
            <div class="message ${alignClass}">
                <div class="meta-info">
                    <span>${displayName}</span>
                    <span>${timeStr}</span>
                </div>
                ${e.content.replace(/(https?:\/\/[^\s]+)/g, u => `<a href="${u}" target="_blank">${u}</a>`)}
                <div class="mood-tag">• ${e.mood}</div>
            </div>`;
        });
        
        feed.innerHTML = html;
        if(wasAtBottom) feed.scrollTop = feed.scrollHeight;
    } catch(e){}
}
async function sendJournal() {
    const txt = document.getElementById('content');
    const auth = nameInput.value.trim();
    if(!auth || !txt.value.trim()) return alert("Set Name First");
    await fetch(API.journal, { method:"POST", body:JSON.stringify({ content:txt.value, mood:document.getElementById('mood').value, author:auth }) });
    txt.value = ""; loadJournal();
}
const chatArea = document.getElementById('content');
if(chatArea) chatArea.addEventListener('keydown', (e)=>{if(e.ctrlKey&&e.key==='Enter')sendJournal()});

// ======================================================
// 5. BUCKET LIST
// ======================================================
async function loadBucket() {
    try {
        const res = await fetch(API.bucket); const data = await res.json();
        const myName = nameInput.value.trim().toLowerCase();
        
        document.getElementById('bucket-list').innerHTML = data.map(item => {
            const isMine = item.author.toLowerCase() === myName;
            const lockedClass = isMine ? '' : 'locked';
            const clickAction = isMine ? `onclick="toggleBucket(${item.id}, ${!item.is_done})"` : '';
            const deleteBtn = isMine ? `<button class="btn-delete" onclick="deleteBucket(event, ${item.id})">✖</button>` : '';
            const checkboxSymbol = isMine ? '✔' : '🔒'; 

            return `
            <div class="bucket-item ${item.is_done ? 'done' : ''} ${lockedClass}">
                <div class="bucket-left" ${clickAction}>
                    <div class="bucket-checkbox">${item.is_done ? '✔' : checkboxSymbol}</div>
                    <div>
                        <span class="bucket-text">${item.text}</span>
                        <div class="bucket-author">by ${item.author}</div>
                    </div>
                </div>
                ${deleteBtn}
            </div>`;
        }).join('');
    } catch(e){}
}
async function addBucket() {
    const txt = document.getElementById('bucketInput'); const auth = nameInput.value.trim();
    if(!txt.value.trim() || !auth) return alert("Fill info");
    await fetch(API.bucket, { method:"POST", body:JSON.stringify({ text: txt.value, author: auth }) });
    txt.value = ""; loadBucket();
}
async function toggleBucket(id, status) { await fetch(API.bucket, { method:"POST", body:JSON.stringify({ id: id, is_done: status }) }); loadBucket(); }
async function deleteBucket(e, id) { e.stopPropagation(); if(!confirm("Delete?")) return; await fetch(API.bucket + "?action=delete", { method:"POST", body:JSON.stringify({ id: id }) }); loadBucket(); }

// ======================================================
// 6. DEEP DIVE
// ======================================================
async function loadDeep() {
    try {
        const res = await fetch(API.deep); const data = await res.json();
        document.getElementById('dailyQuestion').textContent = data.question;
        
        const revealArea = document.getElementById('revealArea');
        const answerArea = document.getElementById('answerArea');
        const hasA = data.answer_a && data.answer_a !== "";
        const hasB = data.answer_b && data.answer_b !== "";

        if(hasA && hasB) {
            answerArea.classList.add('hidden'); revealArea.classList.remove('hidden');
            revealArea.innerHTML = `
                <div class="status-pill complete">✨ BOTH ANSWERED</div>
                <div class="answer-box"><div class="answer-author">${data.author_a}</div><div class="answer-content">"${data.answer_a}"</div></div>
                <div class="answer-box"><div class="answer-author">${data.author_b}</div><div class="answer-content">"${data.answer_b}"</div></div>`;
        } else {
            revealArea.classList.add('hidden'); answerArea.classList.remove('hidden');
            const old = document.querySelector('.status-pill'); if(old) old.remove();
            let st = "";
            if (hasA && !hasB) st = `<div class="status-pill waiting">⏳ Waiting for Partner... (${data.author_a} done)</div>`;
            else if (!hasA && hasB) st = `<div class="status-pill waiting">⏳ Waiting for Partner... (${data.author_b} done)</div>`;
            if(st) document.getElementById('dailyQuestion').insertAdjacentHTML('afterend', st);
        }
    } catch(e){}
}
async function sendDeep() {
    const ans = document.getElementById('deepAnswer'); const auth = nameInput.value.trim();
    if(!auth || !ans.value.trim()) return;
    await fetch(API.deep, { method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({ type:"answer", author:auth, content:ans.value }) });
    ans.value = ""; loadDeep();
}
async function submitNewQuestion() {
    const input = document.getElementById('newQInput'); const q = input.value.trim(); const auth = nameInput.value.trim();
    if(!auth || !q) return;
    await fetch(API.deep, { method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({ type:"new_question", author:auth, content:q }) });
    alert("Question Planted! 🌱"); input.value = "";
}

// ======================================================
// 7. MEMORIES
// ======================================================
async function loadMemories() {
    try {
        const res = await fetch(API.memories); const data = await res.json();
        document.getElementById('memory-grid').innerHTML = data.map(m => `
            <div class="polaroid">
                <img src="${m.image}" loading="lazy">
                <div style="font-weight:bold; margin-bottom:5px;">${m.caption}</div>
                <div style="font-size:0.6rem; color:#888;">${m.date}</div>
            </div>`).join('');
    } catch(e){}
}
function uploadMemory(input) {
    const file = input.files[0]; if(!file || file.size>8*1024*1024) return alert("File too big");
    const reader = new FileReader();
    reader.onload = async (e) => {
        const cap = prompt("Caption:"); if(cap===null) return;
        await fetch(API.memories, { method:"POST", headers:{"Content-Type":"application/json"}, body:JSON.stringify({ image: e.target.result, caption: cap||"Untitled" }) });
        loadMemories();
    };
    reader.readAsDataURL(file);
}

// INIT
switchTab('journal');