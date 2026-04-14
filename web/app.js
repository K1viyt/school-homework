// =========================================================
// API-слой — всё общение с бэкендом в одном месте.
// Любой запрос автоматически получает Bearer-токен.
// =========================================================
const API = ""; // тот же origin, что и фронт

function getToken() { return localStorage.getItem("token"); }
function setAuth(data) {
  localStorage.setItem("token", data.token);
  localStorage.setItem("role", data.role);
  localStorage.setItem("username", data.username);
}
function clearAuth() {
  localStorage.removeItem("token");
  localStorage.removeItem("role");
  localStorage.removeItem("username");
}

async function request(path, { method = "GET", body = null, form = false, json = false } = {}) {
  const headers = {};
  const token = getToken();
  if (token) headers["Authorization"] = "Bearer " + token;
  if (json) headers["Content-Type"] = "application/json";

  const res = await fetch(API + path, { method, headers, body });
  const text = await res.text();
  if (!res.ok) throw new Error(text || ("HTTP " + res.status));
  try { return JSON.parse(text); } catch { return text; }
}

const api = {
  register: (form) => request("/registr", { method: "POST", body: form }),
  login:    (form) => request("/login",   { method: "POST", body: form }),
  logout:   ()     => request("/logout",  { method: "POST" }),
  profile:  ()     => request("/profile"),

  listHomeworks: ()    => request("/homeworks"),
  upload:   (form)     => request("/upload", { method: "POST", body: form }),
  delete:   (id)       => request("/homeworks/delete/" + id, { method: "DELETE" }),
  update:   (id, data) => request("/homeworks/update/" + id, {
    method: "PATCH", body: JSON.stringify(data), json: true,
  }),
  replace:  (id, form) => request("/homeworks/replace/" + id, { method: "PUT", body: form }),
};

// =========================================================
// UI-утилиты
// =========================================================
const $ = (sel) => document.querySelector(sel);
const $$ = (sel) => document.querySelectorAll(sel);

function toast(msg, type = "success") {
  const el = $("#toast");
  el.textContent = msg;
  el.className = "toast " + type;
  el.hidden = false;
  clearTimeout(toast._t);
  toast._t = setTimeout(() => { el.hidden = true; }, 3000);
}

function showSection(role) {
  $("#auth-section").hidden    = role !== null;
  $("#student-section").hidden = role !== "student";
  $("#teacher-section").hidden = role !== "teacher";
  $("#admin-section").hidden   = role !== "admin";
  $("#user-info").hidden       = role === null;
}

function renderUserInfo() {
  $("#user-name").textContent = localStorage.getItem("username") || "";
  $("#user-role").textContent = localStorage.getItem("role") || "";
}

// =========================================================
// Вкладки Вход / Регистрация
// =========================================================
$$(".tab").forEach(btn => {
  btn.addEventListener("click", () => {
    $$(".tab").forEach(b => b.classList.remove("active"));
    $$(".tab-content").forEach(c => c.classList.remove("active"));
    btn.classList.add("active");
    const target = btn.dataset.tab === "login" ? "#login-form" : "#register-form";
    $(target).classList.add("active");
  });
});

// =========================================================
// Аутентификация
// =========================================================
$("#login-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    const fd = new FormData(e.target);
    const data = await api.login(fd);
    setAuth(data);
    toast("Добро пожаловать, " + data.username);
    await bootstrap();
  } catch (err) { toast(err.message, "error"); }
});

$("#register-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    const fd = new FormData(e.target);
    const data = await api.register(fd);
    toast("Регистрация успешна. Ваш username: " + data.username);
    // Автоподстановка username во вкладку логина
    $("#login-form [name=username]").value = data.username;
    $$(".tab")[0].click();
  } catch (err) { toast(err.message, "error"); }
});

$("#logout-btn").addEventListener("click", async () => {
  try { await api.logout(); } catch (_) { /* токен и так чистим */ }
  clearAuth();
  showSection(null);
  toast("Вы вышли");
});

// =========================================================
// Ученик — загрузка и список своих ДЗ
// =========================================================
$("#upload-form").addEventListener("submit", async (e) => {
  e.preventDefault();
  try {
    const fd = new FormData(e.target);
    await api.upload(fd);
    toast("Файл загружен");
    e.target.reset();
    await loadStudentHomeworks();
  } catch (err) { toast(err.message, "error"); }
});

async function loadStudentHomeworks() {
  const list = await api.listHomeworks();
  renderHomeworks($("#student-homeworks"), list || [], /* withActions */ true);
}

// =========================================================
// Учитель — просмотр всех ДЗ
// =========================================================
async function loadTeacherHomeworks() {
  const list = await api.listHomeworks();
  renderHomeworks($("#teacher-homeworks"), list || [], /* withActions */ false);
}

// =========================================================
// Рендер списка ДЗ
// =========================================================
function renderHomeworks(root, items, withActions) {
  root.innerHTML = "";
  if (items.length === 0) {
    root.innerHTML = "<li>Пока ничего нет.</li>";
    return;
  }
  for (const hw of items) {
    const li = document.createElement("li");

    const title = document.createElement("div");
    title.innerHTML = `<strong>${escapeHtml(hw.filename)}</strong>`;
    li.appendChild(title);

    const meta = document.createElement("div");
    meta.className = "homework-meta";
    meta.textContent = [
      hw.subject     ? "Предмет: " + hw.subject         : null,
      hw.description ? "Описание: " + hw.description    : null,
    ].filter(Boolean).join(" · ") || "—";
    li.appendChild(meta);

    if (withActions) {
      const actions = document.createElement("div");
      actions.className = "homework-actions";

      const editBtn = document.createElement("button");
      editBtn.className = "btn-small";
      editBtn.textContent = "Изменить текст";
      editBtn.onclick = () => onEdit(hw);
      actions.appendChild(editBtn);

      const replaceBtn = document.createElement("button");
      replaceBtn.className = "btn-small";
      replaceBtn.textContent = "Заменить файл";
      replaceBtn.onclick = () => onReplace(hw);
      actions.appendChild(replaceBtn);

      const delBtn = document.createElement("button");
      delBtn.className = "btn-small btn-danger";
      delBtn.textContent = "Удалить";
      delBtn.onclick = () => onDelete(hw);
      actions.appendChild(delBtn);

      li.appendChild(actions);
    }

    root.appendChild(li);
  }
}

async function onEdit(hw) {
  const subject = prompt("Новый предмет:", hw.subject || "");
  if (subject === null) return;
  const description = prompt("Новое описание:", hw.description || "");
  if (description === null) return;
  try {
    await api.update(hw.id, { subject, description });
    toast("Обновлено");
    await loadStudentHomeworks();
  } catch (err) { toast(err.message, "error"); }
}

async function onReplace(hw) {
  // Создаём временный input для выбора файла.
  const input = document.createElement("input");
  input.type = "file";
  input.onchange = async () => {
    if (!input.files.length) return;
    const fd = new FormData();
    fd.append("file", input.files[0]);
    fd.append("subject", hw.subject || "");
    fd.append("description", hw.description || "");
    try {
      await api.replace(hw.id, fd);
      toast("Файл заменён");
      await loadStudentHomeworks();
    } catch (err) { toast(err.message, "error"); }
  };
  input.click();
}

async function onDelete(hw) {
  if (!confirm(`Удалить "${hw.filename}"?`)) return;
  try {
    await api.delete(hw.id);
    toast("Удалено");
    await loadStudentHomeworks();
  } catch (err) { toast(err.message, "error"); }
}

function escapeHtml(s) {
  return String(s).replace(/[&<>"']/g, c => ({
    "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;",
  }[c]));
}

// =========================================================
// Старт
// =========================================================
async function bootstrap() {
  const token = getToken();
  if (!token) { showSection(null); return; }

  try {
    // Проверяем токен через /profile и освежаем роль из ответа сервера.
    const prof = await api.profile();
    const role = prof.role || localStorage.getItem("role");
    localStorage.setItem("role", role);
    renderUserInfo();
    showSection(role);

    if (role === "student") await loadStudentHomeworks();
    else if (role === "teacher") await loadTeacherHomeworks();
  } catch (err) {
    // Токен невалиден / истёк — разлогиниваем.
    clearAuth();
    showSection(null);
  }
}

bootstrap();
