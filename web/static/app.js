// Application state
let ws = null;
let currentDocumentId = null;
let userId = generateUserId();
let userName = "";
let isConnected = false;
let saveTimeout = null;

// DOM elements
const editor = document.getElementById("editor");
const preview = document.getElementById("preview");
const status = document.getElementById("status");
const modal = document.getElementById("modal");
const documentIdInput = document.getElementById("documentId");
const userNameInput = document.getElementById("userName");
const joinBtn = document.getElementById("joinBtn");
const newDocBtn = document.getElementById("newDocBtn");
const saveBtn = document.getElementById("saveBtn");
const userDropdownBtn = document.getElementById("userDropdownBtn");
const userDropdownMenu = document.getElementById("userDropdownMenu");
const userCount = document.getElementById("userCount");
const userList = document.getElementById("userList");

// Initialize on page load
window.addEventListener("DOMContentLoaded", () => {
  // Check if document ID is in URL
  const urlParams = new URLSearchParams(window.location.search);
  const docId = urlParams.get("doc");

  if (docId) {
    documentIdInput.value = docId;
  }

  // Show modal to get user name
  showModal();
});

// Event listeners
joinBtn.addEventListener("click", handleJoin);
userNameInput.addEventListener("keypress", (e) => {
  if (e.key === "Enter") handleJoin();
});

newDocBtn.addEventListener("click", createNewDocument);
saveBtn.addEventListener("click", saveDocument);

editor.addEventListener("input", (e) => {
  updatePreview();
  broadcastChange(e.target.value);
  scheduleSave();
});

// User dropdown toggle
userDropdownBtn.addEventListener("click", (e) => {
  e.stopPropagation();
  userDropdownMenu.classList.toggle("hidden");
});

// Close dropdown when clicking outside
document.addEventListener("click", () => {
  userDropdownMenu.classList.add("hidden");
});

// Generate unique user ID
function generateUserId() {
  return "user_" + Math.random().toString(36).substr(2, 9);
}

// Show/hide modal
function showModal() {
  modal.classList.remove("hidden");
}

function hideModal() {
  modal.classList.add("hidden");
}

// Handle joining a document
async function handleJoin() {
  userName = userNameInput.value.trim();

  if (!userName) {
    alert("Please enter your name");
    return;
  }

  let docId = documentIdInput.value.trim();

  // If no document ID provided, check URL or create new
  if (!docId) {
    const urlParams = new URLSearchParams(window.location.search);
    docId = urlParams.get("doc");

    if (!docId) {
      // Create new document
      await createNewDocument();
      return;
    }
  }

  currentDocumentId = docId;

  // Load document from server
  await loadDocument(docId);

  // Connect to websocket
  connectWebSocket(docId);

  // Update URL
  history.pushState({}, "", `?doc=${docId}`);

  hideModal();
}

// Create new document
async function createNewDocument() {
  try {
    const response = await fetch("/documents", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
    });

    if (!response.ok) {
      throw new Error("Failed to create document");
    }

    const doc = await response.json();
    currentDocumentId = doc.id;

    // Update URL
    history.pushState({}, "", `?doc=${doc.id}`);

    // Connect to websocket
    connectWebSocket(doc.id);

    hideModal();

    // Show success message
    showStatus("New document created!", "success");
  } catch (error) {
    console.error("Error creating document:", error);
    alert("Failed to create new document");
  }
}

// Load document from server
async function loadDocument(docId) {
  try {
    const response = await fetch(`/documents/${docId}`);

    if (!response.ok) {
      throw new Error("Document not found");
    }

    const doc = await response.json();
    editor.value = doc.content || "";
    updatePreview();
  } catch (error) {
    console.error("Error loading document:", error);
    alert("Failed to load document");
  }
}

// Save document to server
async function saveDocument() {
  if (!currentDocumentId) return;

  try {
    const response = await fetch(`/documents/${currentDocumentId}`, {
      method: "PUT",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify({
        content: editor.value,
      }),
    });

    if (!response.ok) {
      throw new Error("Failed to save document");
    }

    showStatus("Saved!", "success");
  } catch (error) {
    console.error("Error saving document:", error);
    showStatus("Save failed", "error");
  }
}

// Schedule auto-save
function scheduleSave() {
  if (saveTimeout) {
    clearTimeout(saveTimeout);
  }

  saveTimeout = setTimeout(() => {
    saveDocument();
  }, 2000);
}

// Connect to websocket
function connectWebSocket(docId) {
  const protocol = window.location.protocol === "https:" ? "wss:" : "ws:";
  const wsUrl = `${protocol}//${window.location.host}/ws?documentId=${docId}&userId=${userId}&userName=${encodeURIComponent(userName)}`;

  ws = new WebSocket(wsUrl);

  ws.onopen = () => {
    isConnected = true;
    updateStatus("Connected", "connected");
    console.log("WebSocket connected");
  };

  ws.onmessage = (event) => {
    try {
      const data = JSON.parse(event.data);
      handleWebSocketMessage(data);
    } catch (error) {
      console.error("Error parsing message:", error);
    }
  };

  ws.onerror = (error) => {
    console.error("WebSocket error:", error);
    updateStatus("Error", "disconnected");
  };

  ws.onclose = () => {
    isConnected = false;
    updateStatus("Disconnected", "disconnected");
    console.log("WebSocket disconnected");

    // Attempt to reconnect after 3 seconds
    setTimeout(() => {
      if (currentDocumentId) {
        connectWebSocket(currentDocumentId);
      }
    }, 3000);
  };
}

// Handle incoming websocket messages
function handleWebSocketMessage(data) {
  if (data.type === "content") {
    // Update editor content (but preserve cursor position)
    const start = editor.selectionStart;
    const end = editor.selectionEnd;

    editor.value = data.content;
    updatePreview();

    // Restore cursor position
    editor.setSelectionRange(start, end);
  } else if (data.type === "userList") {
    // Update the active user list
    updateUserList(data.users);
  }
}

// Broadcast content change to other clients
function broadcastChange(content) {
  if (ws && isConnected) {
    ws.send(
      JSON.stringify({
        type: "content",
        content: content,
        userId: userId,
        userName: userName,
      }),
    );
  }
}

// Update preview
function updatePreview() {
  preview.innerHTML = marked.parse(
    editor.value || "# Start typing...\n\nYour markdown will appear here.",
  );
}

// Update connection status
function updateStatus(text, className) {
  status.textContent = text;
  status.className = `status ${className}`;
}

// Show temporary status message
function showStatus(message, type) {
  const originalText = status.textContent;
  const originalClass = status.className;

  status.textContent = message;
  status.className = `status ${type === "success" ? "connected" : "disconnected"}`;

  setTimeout(() => {
    status.textContent = originalText;
    status.className = originalClass;
  }, 2000);
}

// Update user list display
function updateUserList(users) {
  // Update user count
  userCount.textContent = users.length;

  // Clear existing user list
  userList.innerHTML = "";

  if (users.length === 0) {
    userList.innerHTML = '<div class="no-users">No users connected</div>';
    return;
  }

  // Create list items for each user
  users.forEach((user) => {
    const userItem = document.createElement("div");
    userItem.className = "user-list-item";

    // Create avatar with first letter of name
    const avatar = document.createElement("div");
    avatar.className = "user-avatar";
    avatar.textContent = user.userName.charAt(0).toUpperCase();

    // Assign color based on userId for consistency
    const colors = [
      "#3498db",
      "#e74c3c",
      "#2ecc71",
      "#f39c12",
      "#9b59b6",
      "#1abc9c",
    ];
    const colorIndex =
      user.userId.split("").reduce((acc, char) => acc + char.charCodeAt(0), 0) %
      colors.length;
    avatar.style.backgroundColor = colors[colorIndex];

    // Create name element
    const nameElement = document.createElement("div");
    nameElement.className = "user-name";
    nameElement.textContent = user.userName;

    userItem.appendChild(avatar);
    userItem.appendChild(nameElement);
    userList.appendChild(userItem);
  });
}

// Initialize preview
updatePreview();
