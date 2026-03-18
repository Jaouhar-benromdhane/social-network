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

        <div class="connections">
          <article class="connection-block">
            <h3>Followers</h3>
            <ul v-if="profile?.followers?.length" class="mini-list">
              <li v-for="follower in profile.followers" :key="follower.id">
                {{ follower.first_name }} {{ follower.last_name }}
              </li>
            </ul>
            <p v-else class="muted">No followers yet.</p>
          </article>

          <article class="connection-block">
            <h3>Following</h3>
            <ul v-if="profile?.following?.length" class="mini-list">
              <li v-for="followed in profile.following" :key="followed.id">
                {{ followed.first_name }} {{ followed.last_name }}
              </li>
            </ul>
            <p v-else class="muted">You are not following anyone yet.</p>
          </article>
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

        <section class="network">
          <h3>Discover users</h3>
          <p class="muted">Use this section to test follow requests, auto-follow public profiles, accept or decline requests, and unfollow.</p>

          <ul v-if="networkUsers.length" class="user-list">
            <li v-for="user in networkUsers" :key="user.id" class="user-item">
              <div>
                <strong>{{ user.first_name }} {{ user.last_name }}</strong>
                <p class="muted small">{{ user.email }} | {{ user.profile_visibility }}</p>
              </div>

              <div class="user-actions">
                <span v-if="user.is_self" class="pill">You</span>
                <span v-else-if="user.is_following" class="pill success">Following</span>
                <span v-else-if="user.request_status === 'pending'" class="pill pending">Request pending</span>

                <button
                  v-if="!user.is_self && !user.is_following && user.request_status !== 'pending'"
                  type="button"
                  class="tiny"
                  @click="sendFollowRequest(user.id)"
                >
                  {{ user.profile_visibility === 'private' ? 'Send follow request' : 'Follow' }}
                </button>

                <button
                  v-if="!user.is_self && user.is_following"
                  type="button"
                  class="tiny secondary"
                  @click="unfollowUser(user.id)"
                >
                  Unfollow
                </button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No users available yet.</p>
        </section>

        <section class="network">
          <h3>Incoming follow requests</h3>

          <ul v-if="incomingRequests.length" class="user-list">
            <li v-for="request in incomingRequests" :key="request.id" class="user-item">
              <div>
                <strong>{{ request.requester.first_name }} {{ request.requester.last_name }}</strong>
                <p class="muted small">{{ request.requester.email }}</p>
              </div>

              <div class="user-actions">
                <button type="button" class="tiny" @click="respondFollowRequest(request.id, 'accept')">Accept</button>
                <button type="button" class="tiny secondary" @click="respondFollowRequest(request.id, 'decline')">Decline</button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No pending requests.</p>
        </section>

        <section class="network">
          <h3>Profile visibility check</h3>
          <p class="muted">Paste a user ID and test private/public profile access.</p>

          <div class="inline-form">
            <input v-model="profileViewUserID" type="text" placeholder="Target user id" />
            <button type="button" class="tiny" @click="viewOtherProfile">View profile</button>
          </div>

          <p v-if="profileViewError" class="error">{{ profileViewError }}</p>
          <p v-if="profileViewResult" class="muted">
            Profile visible: {{ profileViewResult.user.first_name }} {{ profileViewResult.user.last_name }}
            ({{ profileViewResult.user.profile_visibility }})
            | visible posts: {{ profileViewResult.posts?.length ?? 0 }}
          </p>
        </section>

        <section class="network">
          <h3>Create post</h3>
          <form class="form compact-form" @submit.prevent="submitPost">
            <label>
              Content
              <textarea v-model="postForm.content" rows="3" placeholder="Write something..."></textarea>
            </label>

            <label>
              Privacy
              <select v-model="postForm.privacy">
                <option value="public">Public</option>
                <option value="almost_private">Almost private (followers only)</option>
                <option value="private">Private (selected followers only)</option>
              </select>
            </label>

            <div v-if="postForm.privacy === 'private'" class="private-targets">
              <p class="muted small">Select followers allowed to view this private post.</p>
              <div v-if="followerAudience.length" class="checks">
                <label v-for="follower in followerAudience" :key="follower.id" class="check-item">
                  <input v-model="postForm.allowed_user_ids" type="checkbox" :value="follower.id" />
                  <span>{{ follower.first_name }} {{ follower.last_name }}</span>
                </label>
              </div>
              <p v-else class="muted small">No followers available yet. Private posts need at least one follower.</p>
            </div>

            <label>
              Image / GIF (optional)
              <input type="file" accept="image/jpeg,image/png,image/gif" @change="onPostMediaChange" />
            </label>

            <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Create post' }}</button>
          </form>
        </section>

        <section class="network">
          <h3>Feed</h3>
          <p class="muted">Posts visible to your current account based on privacy rules.</p>

          <ul v-if="posts.length" class="post-list">
            <li v-for="post in posts" :key="post.id" class="post-item">
              <header class="post-head">
                <div>
                  <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
                  <p class="muted small">{{ formatDate(post.created_at) }}</p>
                </div>
                <span class="pill">{{ post.privacy }}</span>
              </header>

              <p v-if="post.content" class="post-content">{{ post.content }}</p>
              <img v-if="post.media_path" class="post-media" :src="post.media_path" alt="post media" />

              <p
                v-if="post.user_id === me.id && post.privacy === 'private' && post.allowed_user_ids?.length"
                class="muted small"
              >
                Allowed user IDs: {{ post.allowed_user_ids.join(', ') }}
              </p>

              <div class="comment-block">
                <h4>Comments ({{ post.comments?.length || 0 }})</h4>

                <ul v-if="post.comments?.length" class="comment-list">
                  <li v-for="comment in post.comments" :key="comment.id" class="comment-item">
                    <p class="small">
                      <strong>{{ comment.author.first_name }} {{ comment.author.last_name }}</strong>
                      · {{ formatDate(comment.created_at) }}
                    </p>
                    <p v-if="comment.content" class="comment-content">{{ comment.content }}</p>
                    <img
                      v-if="comment.media_path"
                      class="comment-media"
                      :src="comment.media_path"
                      alt="comment media"
                    />
                  </li>
                </ul>
                <p v-else class="muted small">No comments yet.</p>

                <form class="comment-form" @submit.prevent="submitComment(post.id)">
                  <textarea
                    v-model="commentDrafts[post.id]"
                    rows="2"
                    placeholder="Write a comment (or upload media)..."
                  ></textarea>
                  <div class="comment-actions">
                    <input
                      type="file"
                      accept="image/jpeg,image/png,image/gif"
                      @change="onCommentMediaChange(post.id, $event)"
                    />
                    <button type="submit" class="tiny" :disabled="loading">Comment</button>
                  </div>
                </form>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No visible posts yet.</p>
        </section>

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
const networkUsers = ref([]);
const incomingRequests = ref([]);
const profileViewUserID = ref("");
const profileViewResult = ref(null);
const profileViewError = ref("");
const posts = ref([]);

const postForm = reactive({
  content: "",
  privacy: "public",
  media: null,
  allowed_user_ids: [],
});

const commentDrafts = reactive({});
const commentMedia = reactive({});

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
const followerAudience = computed(() => profile.value?.followers || []);

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
      networkUsers.value = [];
      incomingRequests.value = [];
      posts.value = [];
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
    await loadNetworkData();
    await loadFeed();
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
    await loadNetworkData();
    await loadFeed();
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
    await loadNetworkData();
    await loadFeed();
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
    await loadFeed();
  } catch (_err) {
    authError.value = "Visibility update failed.";
  } finally {
    loading.value = false;
  }
}

async function loadNetworkData() {
  if (!me.value) {
    networkUsers.value = [];
    incomingRequests.value = [];
    return;
  }

  try {
    const [usersRes, requestsRes] = await Promise.all([
      fetch("/api/users", { credentials: "include" }),
      fetch("/api/follows/requests/incoming", { credentials: "include" }),
    ]);

    if (usersRes.ok) {
      const usersPayload = await usersRes.json();
      networkUsers.value = usersPayload.users || [];
    }

    if (requestsRes.ok) {
      const requestsPayload = await requestsRes.json();
      incomingRequests.value = requestsPayload.requests || [];
    }
  } catch (_err) {
    // Leave existing data as-is to avoid disruptive UI flicker.
  }
}

async function sendFollowRequest(targetUserID) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/follows/request", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ target_user_id: targetUserID }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to follow this user.";
      return;
    }

    await loadMyProfile();
    await loadNetworkData();
    await loadFeed();
  } catch (_err) {
    authError.value = "Follow request failed.";
  } finally {
    loading.value = false;
  }
}

async function unfollowUser(targetUserID) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/follows/unfollow", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ target_user_id: targetUserID }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to unfollow user.";
      return;
    }

    await loadMyProfile();
    await loadNetworkData();
    await loadFeed();
  } catch (_err) {
    authError.value = "Unfollow request failed.";
  } finally {
    loading.value = false;
  }
}

async function respondFollowRequest(requestID, action) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/follows/requests/respond", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        request_id: requestID,
        action,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to process follow request.";
      return;
    }

    await loadMyProfile();
    await loadNetworkData();
    await loadFeed();
  } catch (_err) {
    authError.value = "Failed to respond to request.";
  } finally {
    loading.value = false;
  }
}

async function viewOtherProfile() {
  profileViewError.value = "";
  profileViewResult.value = null;

  const userID = profileViewUserID.value.trim();
  if (!userID) {
    profileViewError.value = "Please provide a user id.";
    return;
  }

  try {
    const response = await fetch(`/api/profile/view?user_id=${encodeURIComponent(userID)}`, {
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      profileViewError.value = payload.error || "Unable to view target profile.";
      return;
    }

    profileViewResult.value = payload;
  } catch (_err) {
    profileViewError.value = "Profile lookup failed.";
  }
}

function onPostMediaChange(event) {
  const [file] = event.target.files || [];
  postForm.media = file || null;
}

function onCommentMediaChange(postID, event) {
  const [file] = event.target.files || [];
  commentMedia[postID] = file || null;
}

async function loadFeed(userID = "") {
  if (!me.value) {
    posts.value = [];
    return;
  }

  const suffix = userID ? `?user_id=${encodeURIComponent(userID)}` : "";
  try {
    const response = await fetch(`/api/posts/feed${suffix}`, {
      credentials: "include",
    });

    if (!response.ok) {
      return;
    }

    const payload = await response.json();
    posts.value = payload.posts || [];
  } catch (_err) {
    // Keep previous feed content if request fails.
  }
}

async function submitPost() {
  loading.value = true;
  authError.value = "";

  try {
    const form = new FormData();
    form.set("content", postForm.content);
    form.set("privacy", postForm.privacy);
    if (postForm.privacy === "private") {
      form.set("allowed_user_ids", JSON.stringify(postForm.allowed_user_ids));
    }
    if (postForm.media) {
      form.set("media", postForm.media);
    }

    const response = await fetch("/api/posts", {
      method: "POST",
      body: form,
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create post.";
      return;
    }

    postForm.content = "";
    postForm.privacy = "public";
    postForm.media = null;
    postForm.allowed_user_ids = [];
    await loadMyProfile();
    await loadFeed();
  } catch (_err) {
    authError.value = "Post creation failed.";
  } finally {
    loading.value = false;
  }
}

async function submitComment(postID) {
  loading.value = true;
  authError.value = "";

  try {
    const form = new FormData();
    form.set("post_id", postID);
    form.set("content", commentDrafts[postID] || "");
    if (commentMedia[postID]) {
      form.set("media", commentMedia[postID]);
    }

    const response = await fetch("/api/posts/comments", {
      method: "POST",
      body: form,
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create comment.";
      return;
    }

    commentDrafts[postID] = "";
    commentMedia[postID] = null;
    await loadFeed();
  } catch (_err) {
    authError.value = "Comment creation failed.";
  } finally {
    loading.value = false;
  }
}

function formatDate(rawValue) {
  if (!rawValue) {
    return "";
  }

  const date = new Date(rawValue);
  if (Number.isNaN(date.getTime())) {
    return rawValue;
  }
  return date.toLocaleString();
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
    networkUsers.value = [];
    incomingRequests.value = [];
    posts.value = [];
    postForm.content = "";
    postForm.privacy = "public";
    postForm.media = null;
    postForm.allowed_user_ids = [];
    profileViewUserID.value = "";
    profileViewResult.value = null;
    profileViewError.value = "";
    loading.value = false;
    loginForm.password = "";
  }
}
</script>
