<template>
  <main class="layout">
    <section class="card">
      <p class="eyebrow">Mandatory Project</p>
      <h1>Social Network</h1>

      <div class="status-row">
        <span class="label">API health:</span>
        <strong :class="statusClass">{{ status }}</strong>
      </div>

      <p v-if="error" class="error">{{ error }}</p>

      <template v-if="!me">
        <div class="tabs">
          <button
            type="button"
            class="tab"
            :class="{ active: mode === 'login' }"
            @click="mode = 'login'"
          >
            Login
          </button>
          <button
            type="button"
            class="tab"
            :class="{ active: mode === 'register' }"
            @click="mode = 'register'"
          >
            Register
          </button>
        </div>

        <form v-if="mode === 'login'" class="form" @submit.prevent="submitLogin">
          <label>
            Email
            <input v-model="loginForm.email" type="email" required autocomplete="email" />
          </label>
          <label>
            Password
            <input
              v-model="loginForm.password"
              type="password"
              required
              minlength="8"
              autocomplete="current-password"
            />
          </label>

          <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Login' }}</button>
        </form>

        <form v-else class="form" @submit.prevent="submitRegister">
          <label>
            Email
            <input v-model="registerForm.email" type="email" required autocomplete="email" />
          </label>
          <label>
            Password
            <input
              v-model="registerForm.password"
              type="password"
              required
              minlength="8"
              autocomplete="new-password"
            />
          </label>

          <div class="grid-2">
            <label>
              First Name
              <input v-model="registerForm.first_name" type="text" required autocomplete="given-name" />
            </label>
            <label>
              Last Name
              <input v-model="registerForm.last_name" type="text" required autocomplete="family-name" />
            </label>
          </div>

          <label>
            Date of Birth
            <input v-model="registerForm.date_of_birth" type="date" required />
          </label>

          <label>
            Avatar / Image (optional)
            <input type="file" accept="image/jpeg,image/png,image/gif" @change="onAvatarChange" />
          </label>

          <label>
            Nickname (optional)
            <input v-model="registerForm.nickname" type="text" />
          </label>

          <label>
            About Me (optional)
            <textarea v-model="registerForm.about_me" rows="3"></textarea>
          </label>

          <label>
            Profile visibility
            <select v-model="registerForm.profile_visibility">
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </label>

          <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Create account' }}</button>
        </form>

        <p v-if="authError" class="error">{{ authError }}</p>
      </template>

      <template v-else>
        <div class="profile-head">
          <img v-if="avatarUrl" class="avatar" :src="avatarUrl" alt="avatar" />
          <div>
            <h2 class="name">{{ me.first_name }} {{ me.last_name }}</h2>
            <p class="subline">{{ me.email }}</p>
            <p v-if="me.nickname" class="subline">@{{ me.nickname }}</p>
          </div>
        </div>

        <div class="profile-grid">
          <p><strong>Date of birth:</strong> {{ me.date_of_birth }}</p>
          <p><strong>Followers:</strong> {{ profile?.stats?.followers ?? 0 }}</p>
          <p><strong>Following:</strong> {{ profile?.stats?.following ?? 0 }}</p>
          <p><strong>Posts:</strong> {{ profile?.posts?.length ?? 0 }}</p>
        </div>

        <p v-if="me.about_me" class="about">{{ me.about_me }}</p>

        <div class="actions">
          <label>
            Profile visibility
            <select v-model="visibility">
              <option value="public">Public</option>
              <option value="private">Private</option>
            </select>
          </label>
          <button type="button" @click="updateVisibility" :disabled="loading">Update visibility</button>
          <button type="button" class="secondary" @click="logout" :disabled="loading">Logout</button>
        </div>

        <p v-if="authError" class="error">{{ authError }}</p>
      </template>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, reactive, ref } from "vue";

const status = ref("checking");
const error = ref("");
const authError = ref("");
const loading = ref(false);
const mode = ref("login");
const me = ref(null);
const profile = ref(null);
const visibility = ref("public");

const loginForm = reactive({
  email: "",
  password: "",
});

const registerForm = reactive({
  email: "",
  password: "",
  first_name: "",
  last_name: "",
  date_of_birth: "",
  nickname: "",
  about_me: "",
  profile_visibility: "public",
  avatar: null,
});

const statusClass = computed(() => {
  if (status.value === "ok") return "ok";
  if (status.value === "degraded") return "warn";
  return "pending";
});

const avatarUrl = computed(() => me.value?.avatar_path || "");

onMounted(async () => {
  await checkHealth();
  await loadCurrentUser();
});

async function checkHealth() {
  try {
    const response = await fetch("/api/health");
    const payload = await response.json();
    status.value = payload.status || "unknown";
  } catch (_err) {
    status.value = "unreachable";
    error.value = "Backend is not reachable yet. Start backend server first.";
  }
}

async function loadCurrentUser() {
  try {
    const response = await fetch("/api/auth/me", {
      credentials: "include",
    });

    if (response.status === 401) {
      me.value = null;
      profile.value = null;
      return;
    }

    if (!response.ok) {
      const payload = await response.json();
      authError.value = payload.error || "Unable to load current user.";
      return;
    }

    const payload = await response.json();
    me.value = payload.user;
    visibility.value = payload.user.profile_visibility;
    await loadMyProfile();
  } catch (_err) {
    authError.value = "Unable to reach authentication service.";
  }
}

async function loadMyProfile() {
  try {
    const response = await fetch("/api/profile/me", {
      credentials: "include",
    });

    if (!response.ok) {
      return;
    }

    profile.value = await response.json();
  } catch (_err) {
    // Keep UI usable even if profile summary fails.
  }
}

async function submitLogin() {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/auth/login", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        email: loginForm.email,
        password: loginForm.password,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Login failed.";
      return;
    }

    me.value = payload.user;
    visibility.value = payload.user.profile_visibility;
    loginForm.password = "";
    await loadMyProfile();
  } catch (_err) {
    authError.value = "Login request failed.";
  } finally {
    loading.value = false;
  }
}

async function submitRegister() {
  loading.value = true;
  authError.value = "";

  try {
    const form = new FormData();
    form.set("email", registerForm.email);
    form.set("password", registerForm.password);
    form.set("first_name", registerForm.first_name);
    form.set("last_name", registerForm.last_name);
    form.set("date_of_birth", registerForm.date_of_birth);
    form.set("profile_visibility", registerForm.profile_visibility);

    if (registerForm.nickname.trim()) {
      form.set("nickname", registerForm.nickname.trim());
    }
    if (registerForm.about_me.trim()) {
      form.set("about_me", registerForm.about_me.trim());
    }
    if (registerForm.avatar) {
      form.set("avatar", registerForm.avatar);
    }

    const response = await fetch("/api/auth/register", {
      method: "POST",
      body: form,
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Registration failed.";
      return;
    }

    me.value = payload.user;
    visibility.value = payload.user.profile_visibility;
    await loadMyProfile();
  } catch (_err) {
    authError.value = "Registration request failed.";
  } finally {
    loading.value = false;
  }
}

function onAvatarChange(event) {
  const [file] = event.target.files || [];
  registerForm.avatar = file || null;
}

async function updateVisibility() {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/profile/me/visibility", {
      method: "PATCH",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ visibility: visibility.value }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Failed to update visibility.";
      return;
    }

    me.value = payload.user;
    await loadMyProfile();
  } catch (_err) {
    authError.value = "Visibility update failed.";
  } finally {
    loading.value = false;
  }
}

async function logout() {
  loading.value = true;
  authError.value = "";

  try {
    await fetch("/api/auth/logout", {
      method: "POST",
      credentials: "include",
    });
  } finally {
    me.value = null;
    profile.value = null;
    loading.value = false;
    loginForm.password = "";
  }
}
</script>
