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

        <nav class="page-nav" aria-label="App pages">
          <button type="button" class="tab" :class="{ active: activePage === 'profile' }" @click="setActivePage('profile')">
            Profile
          </button>
          <button type="button" class="tab" :class="{ active: activePage === 'posts' }" @click="setActivePage('posts')">
            Posts
          </button>
          <button type="button" class="tab" :class="{ active: activePage === 'groups' }" @click="setActivePage('groups')">
            Groups
          </button>
          <button type="button" class="tab" :class="{ active: activePage === 'chat' }" @click="setActivePage('chat')">
            Chat & Notifications
          </button>
        </nav>

        <section v-if="activePage === 'profile'" class="network">
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

                <button
                  v-if="!user.is_self && (user.profile_visibility === 'public' || user.is_following)"
                  type="button"
                  class="tiny secondary"
                  @click="viewProfileFromUserCard(user.id)"
                >
                  View profile
                </button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No users available yet.</p>
        </section>

        <section v-if="activePage === 'profile'" class="network">
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

        <section v-if="activePage === 'profile'" class="network">
          <h3>Profile preview</h3>
          <p class="muted">Paste a user ID to open a full preview of that profile.</p>

          <div class="inline-form">
            <input v-model="profileViewUserID" type="text" placeholder="Target user id" />
            <button type="button" class="tiny" @click="viewOtherProfile">View profile</button>
          </div>

          <p v-if="profileViewError" class="error">{{ profileViewError }}</p>

          <article v-if="profileViewResult" class="profile-preview">
            <header class="profile-head">
              <img v-if="profileViewResult.user.avatar_path" class="avatar" :src="profileViewResult.user.avatar_path" alt="profile avatar" />
              <div>
                <h3 class="name">{{ profileViewResult.user.first_name }} {{ profileViewResult.user.last_name }}</h3>
                <p class="subline">{{ profileViewResult.user.email }}</p>
                <p v-if="profileViewResult.user.nickname" class="subline">@{{ profileViewResult.user.nickname }}</p>
              </div>
            </header>

            <div class="profile-grid">
              <p><strong>Visibility:</strong> {{ profileViewResult.user.profile_visibility }}</p>
              <p><strong>Followers:</strong> {{ profileViewResult.stats?.followers ?? 0 }}</p>
              <p><strong>Following:</strong> {{ profileViewResult.stats?.following ?? 0 }}</p>
              <p><strong>Visible posts:</strong> {{ profileViewResult.posts?.length ?? 0 }}</p>
            </div>

            <p v-if="profileViewResult.user.about_me" class="about">{{ profileViewResult.user.about_me }}</p>

            <div class="connections">
              <article class="connection-block">
                <h3>Followers</h3>
                <ul v-if="profileViewResult.followers?.length" class="mini-list">
                  <li v-for="follower in profileViewResult.followers" :key="follower.id">
                    {{ follower.first_name }} {{ follower.last_name }}
                  </li>
                </ul>
                <p v-else class="muted">No followers yet.</p>
              </article>

              <article class="connection-block">
                <h3>Following</h3>
                <ul v-if="profileViewResult.following?.length" class="mini-list">
                  <li v-for="followed in profileViewResult.following" :key="followed.id">
                    {{ followed.first_name }} {{ followed.last_name }}
                  </li>
                </ul>
                <p v-else class="muted">Not following anyone yet.</p>
              </article>
            </div>

            <div class="preview-posts">
              <h4>Visible posts</h4>

              <ul v-if="profileViewResult.posts?.length" class="post-list">
                <li v-for="post in profileViewResult.posts" :key="post.id" class="post-item">
                  <header class="post-head">
                    <div>
                      <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
                      <p class="muted small">{{ formatDate(post.created_at) }}</p>
                    </div>
                    <span class="pill">{{ post.privacy }}</span>
                  </header>

                  <p v-if="post.content" class="post-content">{{ post.content }}</p>
                  <img v-if="post.media_path" class="post-media" :src="post.media_path" alt="post media" />
                </li>
              </ul>
              <p v-else class="muted">No visible posts for this profile.</p>
            </div>
          </article>
        </section>

        <section v-if="activePage === 'posts'" class="network">
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

        <section v-if="activePage === 'posts'" class="network">
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

        <section v-if="activePage === 'groups'" class="network">
          <h3>Groups</h3>

          <form class="form compact-form" @submit.prevent="submitCreateGroup">
            <label>
              Group title
              <input v-model="groupForm.title" type="text" required placeholder="Weekend football" />
            </label>
            <label>
              Group description
              <textarea
                v-model="groupForm.description"
                rows="3"
                required
                placeholder="Group purpose and rules"
              ></textarea>
            </label>

            <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Create group' }}</button>
          </form>

          <div class="group-columns">
            <article class="connection-block">
              <h3>My groups</h3>

              <ul v-if="memberGroups.length" class="user-list">
                <li v-for="group in memberGroups" :key="group.id" class="user-item group-item">
                  <div>
                    <strong>{{ group.title }}</strong>
                    <p class="muted small">{{ group.description }}</p>
                    <p class="muted small">Members: {{ group.member_count }} | Role: {{ group.member_role || 'member' }}</p>
                  </div>

                  <div class="user-actions">
                    <select v-model="groupInviteTargets[group.id]" class="tiny-select">
                      <option value="">Invite follower...</option>
                      <option v-for="follower in followerAudience" :key="follower.id" :value="follower.id">
                        {{ follower.first_name }} {{ follower.last_name }}
                      </option>
                    </select>
                    <button type="button" class="tiny" @click="inviteFollowerToGroup(group.id)">Invite</button>
                  </div>
                </li>
              </ul>
              <p v-else class="muted">You are not a member of any group yet.</p>
            </article>

            <article class="connection-block">
              <h3>Discover groups</h3>

              <ul v-if="discoverGroups.length" class="user-list">
                <li v-for="group in discoverGroups" :key="group.id" class="user-item group-item">
                  <div>
                    <strong>{{ group.title }}</strong>
                    <p class="muted small">{{ group.description }}</p>
                    <p class="muted small">Members: {{ group.member_count }}</p>
                  </div>

                  <div class="user-actions">
                    <span v-if="group.has_pending_invite" class="pill">Invitation pending</span>
                    <span v-else-if="group.has_pending_join_request" class="pill pending">Join request pending</span>
                    <button
                      v-else
                      type="button"
                      class="tiny"
                      @click="requestJoinGroup(group.id)"
                    >
                      Request join
                    </button>
                  </div>
                </li>
              </ul>
              <p v-else class="muted">No public groups to discover right now.</p>
            </article>
          </div>
        </section>

        <section v-if="activePage === 'groups'" class="network">
          <h3>Selected Group Activity</h3>

          <label>
            Group
            <select v-model="selectedGroupID">
              <option value="">Choose one of your groups...</option>
              <option v-for="group in memberGroups" :key="group.id" :value="group.id">
                {{ group.title }}
              </option>
            </select>
          </label>

          <template v-if="selectedGroupID">
            <div class="group-columns">
              <article class="connection-block">
                <h3>Create group post</h3>

                <form class="form compact-form" @submit.prevent="submitGroupPost">
                  <label>
                    Content
                    <textarea
                      v-model="groupPostForm.content"
                      rows="3"
                      placeholder="Share something with your group..."
                    ></textarea>
                  </label>
                  <label>
                    Image / GIF (optional)
                    <input type="file" accept="image/jpeg,image/png,image/gif" @change="onGroupPostMediaChange" />
                  </label>
                  <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Create group post' }}</button>
                </form>

                <h3>Group posts</h3>
                <ul v-if="groupPosts.length" class="post-list">
                  <li v-for="post in groupPosts" :key="post.id" class="post-item">
                    <header class="post-head">
                      <div>
                        <strong>{{ post.author.first_name }} {{ post.author.last_name }}</strong>
                        <p class="muted small">{{ formatDate(post.created_at) }}</p>
                      </div>
                    </header>

                    <p v-if="post.content" class="post-content">{{ post.content }}</p>
                    <img v-if="post.media_path" class="post-media" :src="post.media_path" alt="group post media" />

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
                            alt="group comment media"
                          />
                        </li>
                      </ul>
                      <p v-else class="muted small">No comments yet.</p>

                      <form class="comment-form" @submit.prevent="submitGroupComment(post.id)">
                        <textarea
                          v-model="groupCommentDrafts[post.id]"
                          rows="2"
                          placeholder="Write a group comment (or upload media)..."
                        ></textarea>
                        <div class="comment-actions">
                          <input
                            type="file"
                            accept="image/jpeg,image/png,image/gif"
                            @change="onGroupCommentMediaChange(post.id, $event)"
                          />
                          <button type="submit" class="tiny" :disabled="loading">Comment</button>
                        </div>
                      </form>
                    </div>
                  </li>
                </ul>
                <p v-else class="muted">No group posts yet.</p>
              </article>

              <article class="connection-block">
                <h3>Create group event</h3>

                <form class="form compact-form" @submit.prevent="submitGroupEvent">
                  <label>
                    Title
                    <input v-model="groupEventForm.title" type="text" required placeholder="Friday meetup" />
                  </label>
                  <label>
                    Description
                    <textarea
                      v-model="groupEventForm.description"
                      rows="2"
                      required
                      placeholder="Where and what is planned"
                    ></textarea>
                  </label>
                  <label>
                    Day / Time
                    <input v-model="groupEventForm.event_datetime" type="datetime-local" required />
                  </label>
                  <label>
                    Option 1
                    <input v-model="groupEventForm.option_one" type="text" required placeholder="Going" />
                  </label>
                  <label>
                    Option 2
                    <input v-model="groupEventForm.option_two" type="text" required placeholder="Not going" />
                  </label>
                  <label>
                    Option 3 (optional)
                    <input v-model="groupEventForm.option_three" type="text" placeholder="Maybe" />
                  </label>
                  <button type="submit" :disabled="loading">{{ loading ? 'Please wait...' : 'Create event' }}</button>
                </form>

                <h3>Group events</h3>
                <ul v-if="groupEvents.length" class="user-list">
                  <li v-for="event in groupEvents" :key="event.id" class="user-item group-item">
                    <div>
                      <strong>{{ event.title }}</strong>
                      <p class="muted small">{{ event.description }}</p>
                      <p class="muted small">{{ formatDate(event.event_datetime) }}</p>
                    </div>

                    <div class="group-options">
                      <button
                        v-for="option in event.options"
                        :key="option.id"
                        type="button"
                        class="tiny"
                        :class="{ secondary: event.my_vote_option_id === option.id }"
                        @click="voteGroupEvent(event.id, option.id)"
                      >
                        {{ option.label }} ({{ option.vote_count }})
                      </button>
                    </div>
                  </li>
                </ul>
                <p v-else class="muted">No events yet.</p>
              </article>
            </div>
          </template>

          <p v-else class="muted">Join or create a group to manage group posts and events.</p>
        </section>

        <section v-if="activePage === 'groups'" class="network">
          <h3>Incoming group invites</h3>

          <ul v-if="groupInvitesIncoming.length" class="user-list">
            <li v-for="invite in groupInvitesIncoming" :key="invite.id" class="user-item">
              <div>
                <strong>{{ invite.group.title }}</strong>
                <p class="muted small">
                  Invited by {{ invite.inviter.first_name }} {{ invite.inviter.last_name }}
                </p>
              </div>

              <div class="user-actions">
                <button type="button" class="tiny" @click="respondGroupInvite(invite.id, 'accept')">Accept</button>
                <button type="button" class="tiny secondary" @click="respondGroupInvite(invite.id, 'decline')">Decline</button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No pending group invites.</p>
        </section>

        <section v-if="activePage === 'groups'" class="network">
          <h3>Incoming join requests (creator)</h3>

          <ul v-if="groupJoinRequestsIncoming.length" class="user-list">
            <li v-for="request in groupJoinRequestsIncoming" :key="request.id" class="user-item">
              <div>
                <strong>{{ request.group.title }}</strong>
                <p class="muted small">
                  {{ request.requester.first_name }} {{ request.requester.last_name }} wants to join
                </p>
              </div>

              <div class="user-actions">
                <button type="button" class="tiny" @click="respondGroupJoinRequest(request.id, 'accept')">Accept</button>
                <button
                  type="button"
                  class="tiny secondary"
                  @click="respondGroupJoinRequest(request.id, 'decline')"
                >
                  Decline
                </button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No pending join requests for your groups.</p>
        </section>

        <section v-if="activePage === 'chat'" class="network">
          <h3>Notifications</h3>

          <div class="user-actions">
            <span v-if="notificationsUnread" class="pill pending">Unread: {{ notificationsUnread }}</span>
            <span v-else class="pill success">All read</span>
            <button
              type="button"
              class="tiny secondary"
              :disabled="loading || !notifications.length"
              @click="markAllNotificationsRead"
            >
              Mark all as read
            </button>
          </div>

          <ul v-if="notifications.length" class="user-list">
            <li v-for="notification in notifications" :key="notification.id" class="user-item notification-item">
              <div>
                <strong>{{ formatNotification(notification) }}</strong>
                <p class="muted small">{{ formatDate(notification.created_at) }}</p>
              </div>

              <div class="user-actions">
                <span v-if="notification.is_read" class="pill success">Read</span>
                <button
                  v-else
                  type="button"
                  class="tiny"
                  :disabled="loading"
                  @click="markNotificationRead(notification.id)"
                >
                  Mark read
                </button>
              </div>
            </li>
          </ul>
          <p v-else class="muted">No notifications yet.</p>
        </section>

        <section v-if="activePage === 'chat'" class="network">
          <h3>Private chat</h3>
          <p class="muted">Private chat is enabled only when both users follow each other.</p>

          <label>
            Chat with
            <select v-model="privateChatTargetID">
              <option value="">Choose a mutual follower...</option>
              <option v-for="user in mutualChatUsers" :key="user.id" :value="user.id">
                {{ user.first_name }} {{ user.last_name }}
              </option>
            </select>
          </label>

          <template v-if="privateChatTargetID">
            <ul v-if="privateMessages.length" class="chat-list">
              <li
                v-for="message in privateMessages"
                :key="message.id"
                class="chat-item"
                :class="{ mine: message.sender_id === me.id }"
              >
                <p class="chat-meta">
                  {{ message.sender.first_name }} {{ message.sender.last_name }}
                  · {{ formatDate(message.created_at) }}
                </p>
                <p class="chat-text">{{ message.content }}</p>
              </li>
            </ul>
            <p v-else class="muted">No private messages yet.</p>

            <form class="chat-input" @submit.prevent="sendPrivateMessage">
              <textarea
                v-model="privateMessageDraft"
                rows="2"
                placeholder="Write a private message (emoji OK)"
              ></textarea>
              <button type="submit" class="tiny" :disabled="loading">Send</button>
            </form>
          </template>

          <p v-else class="muted">Choose one mutual follower to start chatting.</p>
        </section>

        <section v-if="activePage === 'chat'" class="network">
          <h3>Group chat</h3>

          <label>
            Group room
            <select v-model="groupChatTargetID">
              <option value="">Choose one of your groups...</option>
              <option v-for="group in memberGroups" :key="group.id" :value="group.id">
                {{ group.title }}
              </option>
            </select>
          </label>

          <template v-if="groupChatTargetID">
            <ul v-if="groupMessages.length" class="chat-list">
              <li
                v-for="message in groupMessages"
                :key="message.id"
                class="chat-item"
                :class="{ mine: message.sender_id === me.id }"
              >
                <p class="chat-meta">
                  {{ message.sender.first_name }} {{ message.sender.last_name }}
                  · {{ formatDate(message.created_at) }}
                </p>
                <p class="chat-text">{{ message.content }}</p>
              </li>
            </ul>
            <p v-else class="muted">No group messages yet.</p>

            <form class="chat-input" @submit.prevent="sendGroupMessage">
              <textarea
                v-model="groupMessageDraft"
                rows="2"
                placeholder="Write a group message (emoji OK)"
              ></textarea>
              <button type="submit" class="tiny" :disabled="loading">Send</button>
            </form>
          </template>

          <p v-else class="muted">Choose one group to open its chat room.</p>
        </section>

        <p v-if="authError" class="error">{{ authError }}</p>
      </template>
    </section>
  </main>
</template>

<script setup>
import { computed, onMounted, onUnmounted, reactive, ref, watch } from "vue";

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
const groups = ref([]);
const groupInvitesIncoming = ref([]);
const groupJoinRequestsIncoming = ref([]);
const selectedGroupID = ref("");
const groupPosts = ref([]);
const groupEvents = ref([]);
const notifications = ref([]);
const notificationsUnread = ref(0);
const privateChatTargetID = ref("");
const privateMessages = ref([]);
const privateMessageDraft = ref("");
const groupChatTargetID = ref("");
const groupMessages = ref([]);
const groupMessageDraft = ref("");
const activePage = ref("profile");

const appPages = new Set(["profile", "posts", "groups", "chat"]);

function setActivePage(page) {
  if (!appPages.has(page)) {
    return;
  }
  activePage.value = page;
}

function syncPageFromHash() {
  const hash = window.location.hash || "";
  const page = hash.startsWith("#/") ? hash.slice(2) : "";
  if (appPages.has(page)) {
    activePage.value = page;
  }
}

function onHashChange() {
  syncPageFromHash();
}

let realtimeSocket = null;
let reconnectTimer = null;
let activityPollTimer = null;

const postForm = reactive({
  content: "",
  privacy: "public",
  media: null,
  allowed_user_ids: [],
});

const groupPostForm = reactive({
  content: "",
  media: null,
});

const groupForm = reactive({
  title: "",
  description: "",
});

const groupEventForm = reactive({
  title: "",
  description: "",
  event_datetime: "",
  option_one: "Going",
  option_two: "Not going",
  option_three: "",
});

const commentDrafts = reactive({});
const commentMedia = reactive({});
const groupInviteTargets = reactive({});
const groupCommentDrafts = reactive({});
const groupCommentMedia = reactive({});

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
const memberGroups = computed(() => groups.value.filter((group) => group.is_member));
const discoverGroups = computed(() => groups.value.filter((group) => !group.is_member));
const mutualChatUsers = computed(() => {
  const followers = profile.value?.followers || [];
  const following = profile.value?.following || [];
  const followingIDs = new Set(following.map((user) => user.id));
  return followers.filter((user) => followingIDs.has(user.id));
});

onMounted(async () => {
  syncPageFromHash();
  window.addEventListener("hashchange", onHashChange);
  await checkHealth();
  await loadCurrentUser();
});

onUnmounted(() => {
  stopActivityPolling();
  window.removeEventListener("hashchange", onHashChange);
  closeRealtimeSocket();
});

watch(activePage, (page) => {
  const nextHash = `#/${page}`;
  if (window.location.hash !== nextHash) {
    history.replaceState(null, "", nextHash);
  }
});

watch(selectedGroupID, async () => {
  await loadSelectedGroupContent();
});

watch(privateChatTargetID, async () => {
  await loadPrivateMessages();
});

watch(groupChatTargetID, async () => {
  await loadGroupMessages();
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
      stopActivityPolling();
      closeRealtimeSocket();
      me.value = null;
      profile.value = null;
      networkUsers.value = [];
      incomingRequests.value = [];
      posts.value = [];
      groups.value = [];
      groupInvitesIncoming.value = [];
      groupJoinRequestsIncoming.value = [];
      selectedGroupID.value = "";
      groupPosts.value = [];
      groupEvents.value = [];
      notifications.value = [];
      notificationsUnread.value = 0;
      privateChatTargetID.value = "";
      privateMessages.value = [];
      privateMessageDraft.value = "";
      groupChatTargetID.value = "";
      groupMessages.value = [];
      groupMessageDraft.value = "";
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
    await loadGroupsData();
    await loadNotifications();
    connectRealtimeSocket();
    startActivityPolling();
  } catch (_err) {
    authError.value = "Unable to reach authentication service.";
  }
}

function startActivityPolling() {
  stopActivityPolling();

  if (!me.value) {
    return;
  }

  activityPollTimer = setInterval(async () => {
    if (!me.value) {
      return;
    }

    if (activePage.value === "profile") {
      await Promise.allSettled([loadNetworkData(), loadMyProfile()]);
      return;
    }

    if (activePage.value === "posts") {
      await loadFeed();
      return;
    }

    if (activePage.value === "groups") {
      await loadGroupsData();
      return;
    }

    if (activePage.value === "chat") {
      await Promise.allSettled([loadNotifications(), loadPrivateMessages(), loadGroupMessages()]);
    }
  }, 5000);
}

function stopActivityPolling() {
  if (!activityPollTimer) {
    return;
  }

  clearInterval(activityPollTimer);
  activityPollTimer = null;
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

    const mutualIDs = new Set(mutualChatUsers.value.map((user) => user.id));
    if (!mutualIDs.has(privateChatTargetID.value)) {
      privateChatTargetID.value = "";
      privateMessages.value = [];
    }
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
    await loadGroupsData();
    await loadNotifications();
    connectRealtimeSocket();
    startActivityPolling();
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
    await loadGroupsData();
    await loadNotifications();
    connectRealtimeSocket();
    startActivityPolling();
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

async function viewProfileFromUserCard(userID) {
  profileViewUserID.value = userID;
  await viewOtherProfile();
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

async function loadGroupsData() {
  if (!me.value) {
    groups.value = [];
    groupInvitesIncoming.value = [];
    groupJoinRequestsIncoming.value = [];
    selectedGroupID.value = "";
    groupPosts.value = [];
    groupEvents.value = [];
    groupChatTargetID.value = "";
    groupMessages.value = [];
    return;
  }

  const previousSelectedGroupID = selectedGroupID.value;
  const previousGroupChatTargetID = groupChatTargetID.value;

  try {
    const [groupsRes, invitesRes, joinRequestsRes] = await Promise.all([
      fetch("/api/groups", { credentials: "include" }),
      fetch("/api/groups/invites/incoming", { credentials: "include" }),
      fetch("/api/groups/requests/incoming", { credentials: "include" }),
    ]);

    if (groupsRes.ok) {
      const groupsPayload = await groupsRes.json();
      groups.value = groupsPayload.groups || [];

      const memberIDs = groups.value.filter((group) => group.is_member).map((group) => group.id);
      if (!memberIDs.includes(selectedGroupID.value)) {
        selectedGroupID.value = memberIDs[0] || "";
      }
      if (!memberIDs.includes(groupChatTargetID.value)) {
        groupChatTargetID.value = memberIDs[0] || "";
      }
    }

    if (invitesRes.ok) {
      const invitesPayload = await invitesRes.json();
      groupInvitesIncoming.value = invitesPayload.invites || [];
    }

    if (joinRequestsRes.ok) {
      const joinRequestsPayload = await joinRequestsRes.json();
      groupJoinRequestsIncoming.value = joinRequestsPayload.requests || [];
    }

    if (!selectedGroupID.value) {
      groupPosts.value = [];
      groupEvents.value = [];
      return;
    }

    if (selectedGroupID.value === previousSelectedGroupID) {
      await loadSelectedGroupContent();
    }

    if (!groupChatTargetID.value) {
      groupMessages.value = [];
      return;
    }

    if (groupChatTargetID.value === previousGroupChatTargetID) {
      await loadGroupMessages();
    }
  } catch (_err) {
    // Keep last loaded values to avoid UI flicker.
  }
}

function onGroupPostMediaChange(event) {
  const [file] = event.target.files || [];
  groupPostForm.media = file || null;
}

function onGroupCommentMediaChange(groupPostID, event) {
  const [file] = event.target.files || [];
  groupCommentMedia[groupPostID] = file || null;
}

async function loadSelectedGroupContent() {
  const groupID = selectedGroupID.value;
  if (!groupID) {
    groupPosts.value = [];
    groupEvents.value = [];
    return;
  }

  try {
    const [postsRes, eventsRes] = await Promise.all([
      fetch(`/api/groups/posts?group_id=${encodeURIComponent(groupID)}`, { credentials: "include" }),
      fetch(`/api/groups/events?group_id=${encodeURIComponent(groupID)}`, { credentials: "include" }),
    ]);

    if (postsRes.ok) {
      const postsPayload = await postsRes.json();
      groupPosts.value = postsPayload.posts || [];
    }

    if (eventsRes.ok) {
      const eventsPayload = await eventsRes.json();
      groupEvents.value = eventsPayload.events || [];
    }
  } catch (_err) {
    // Keep last successful content rendering.
  }
}

async function submitGroupPost() {
  const groupID = selectedGroupID.value;
  if (!groupID) {
    authError.value = "Select a group first.";
    return;
  }

  loading.value = true;
  authError.value = "";

  try {
    const form = new FormData();
    form.set("group_id", groupID);
    form.set("content", groupPostForm.content || "");
    if (groupPostForm.media) {
      form.set("media", groupPostForm.media);
    }

    const response = await fetch("/api/groups/posts", {
      method: "POST",
      body: form,
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create group post.";
      return;
    }

    groupPostForm.content = "";
    groupPostForm.media = null;
    await loadSelectedGroupContent();
  } catch (_err) {
    authError.value = "Group post creation failed.";
  } finally {
    loading.value = false;
  }
}

async function submitGroupComment(groupPostID) {
  loading.value = true;
  authError.value = "";

  try {
    const form = new FormData();
    form.set("group_post_id", groupPostID);
    form.set("content", groupCommentDrafts[groupPostID] || "");
    if (groupCommentMedia[groupPostID]) {
      form.set("media", groupCommentMedia[groupPostID]);
    }

    const response = await fetch("/api/groups/posts/comments", {
      method: "POST",
      body: form,
      credentials: "include",
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create group comment.";
      return;
    }

    groupCommentDrafts[groupPostID] = "";
    groupCommentMedia[groupPostID] = null;
    await loadSelectedGroupContent();
  } catch (_err) {
    authError.value = "Group comment creation failed.";
  } finally {
    loading.value = false;
  }
}

async function submitGroupEvent() {
  const groupID = selectedGroupID.value;
  if (!groupID) {
    authError.value = "Select a group first.";
    return;
  }

  loading.value = true;
  authError.value = "";

  try {
    const options = [
      groupEventForm.option_one,
      groupEventForm.option_two,
      groupEventForm.option_three,
    ]
      .map((option) => option.trim())
      .filter((option) => option.length > 0);

    const response = await fetch("/api/groups/events", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        group_id: groupID,
        title: groupEventForm.title,
        description: groupEventForm.description,
        event_datetime: groupEventForm.event_datetime,
        options,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create event.";
      return;
    }

    groupEventForm.title = "";
    groupEventForm.description = "";
    groupEventForm.event_datetime = "";
    groupEventForm.option_one = "Going";
    groupEventForm.option_two = "Not going";
    groupEventForm.option_three = "";
    await loadSelectedGroupContent();
  } catch (_err) {
    authError.value = "Event creation failed.";
  } finally {
    loading.value = false;
  }
}

async function voteGroupEvent(eventID, optionID) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups/events/vote", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        event_id: eventID,
        option_id: optionID,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to vote for this option.";
      return;
    }

    await loadSelectedGroupContent();
  } catch (_err) {
    authError.value = "Event vote failed.";
  } finally {
    loading.value = false;
  }
}

async function submitCreateGroup() {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        title: groupForm.title,
        description: groupForm.description,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to create group.";
      return;
    }

    groupForm.title = "";
    groupForm.description = "";
    await loadGroupsData();
  } catch (_err) {
    authError.value = "Group creation failed.";
  } finally {
    loading.value = false;
  }
}

async function inviteFollowerToGroup(groupID) {
  const inviteeID = (groupInviteTargets[groupID] || "").trim();
  if (!inviteeID) {
    authError.value = "Select a follower before inviting.";
    return;
  }

  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups/invites", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        group_id: groupID,
        invitee_id: inviteeID,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to invite this user.";
      return;
    }

    groupInviteTargets[groupID] = "";
    await loadGroupsData();
  } catch (_err) {
    authError.value = "Group invitation failed.";
  } finally {
    loading.value = false;
  }
}

async function requestJoinGroup(groupID) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups/requests/join", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({ group_id: groupID }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to request group join.";
      return;
    }

    await loadGroupsData();
  } catch (_err) {
    authError.value = "Join request failed.";
  } finally {
    loading.value = false;
  }
}

async function respondGroupInvite(requestID, action) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups/invites/respond", {
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
      authError.value = payload.error || "Unable to process group invite.";
      return;
    }

    await loadGroupsData();
  } catch (_err) {
    authError.value = "Failed to respond to group invite.";
  } finally {
    loading.value = false;
  }
}

async function respondGroupJoinRequest(requestID, action) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/groups/requests/respond", {
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
      authError.value = payload.error || "Unable to process join request.";
      return;
    }

    await loadGroupsData();
  } catch (_err) {
    authError.value = "Failed to respond to join request.";
  } finally {
    loading.value = false;
  }
}

async function loadNotifications() {
  if (!me.value) {
    notifications.value = [];
    notificationsUnread.value = 0;
    return;
  }

  try {
    const response = await fetch("/api/notifications", {
      credentials: "include",
    });

    if (!response.ok) {
      return;
    }

    const payload = await response.json();
    notifications.value = payload.notifications || [];
    notificationsUnread.value = payload.unread_count || 0;
  } catch (_err) {
    // Keep last notifications snapshot on transient failures.
  }
}

async function markNotificationRead(notificationID) {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/notifications/read", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        notification_id: notificationID,
        read_all: false,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to mark this notification as read.";
      return;
    }

    notifications.value = notifications.value.map((notification) => {
      if (notification.id !== notificationID) {
        return notification;
      }
      return {
        ...notification,
        is_read: true,
      };
    });
    notificationsUnread.value = payload.unread_count || 0;
  } catch (_err) {
    authError.value = "Notification update failed.";
  } finally {
    loading.value = false;
  }
}

async function markAllNotificationsRead() {
  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/notifications/read", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        read_all: true,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to mark notifications as read.";
      return;
    }

    notifications.value = notifications.value.map((notification) => ({
      ...notification,
      is_read: true,
    }));
    notificationsUnread.value = payload.unread_count || 0;
  } catch (_err) {
    authError.value = "Notifications update failed.";
  } finally {
    loading.value = false;
  }
}

async function loadPrivateMessages() {
  if (!me.value || !privateChatTargetID.value) {
    privateMessages.value = [];
    return;
  }

  try {
    const response = await fetch(`/api/chat/private/messages?user_id=${encodeURIComponent(privateChatTargetID.value)}`, {
      credentials: "include",
    });

    if (!response.ok) {
      privateMessages.value = [];
      return;
    }

    const payload = await response.json();
    privateMessages.value = payload.messages || [];
  } catch (_err) {
    // Keep previous messages to avoid hard UI resets.
  }
}

async function sendPrivateMessage() {
  const targetID = privateChatTargetID.value;
  const content = privateMessageDraft.value.trim();

  if (!targetID) {
    authError.value = "Choose a private chat target first.";
    return;
  }
  if (!content) {
    authError.value = "Message content is required.";
    return;
  }

  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/chat/private/messages", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        recipient_id: targetID,
        content,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to send private message.";
      return;
    }

    const sentMessage = payload.message;
    if (sentMessage && !privateMessages.value.some((message) => message.id === sentMessage.id)) {
      privateMessages.value = [...privateMessages.value, sentMessage];
    }
    privateMessageDraft.value = "";
  } catch (_err) {
    authError.value = "Private message sending failed.";
  } finally {
    loading.value = false;
  }
}

async function loadGroupMessages() {
  if (!me.value || !groupChatTargetID.value) {
    groupMessages.value = [];
    return;
  }

  try {
    const response = await fetch(`/api/chat/groups/messages?group_id=${encodeURIComponent(groupChatTargetID.value)}`, {
      credentials: "include",
    });

    if (!response.ok) {
      groupMessages.value = [];
      return;
    }

    const payload = await response.json();
    groupMessages.value = payload.messages || [];
  } catch (_err) {
    // Keep previous messages to avoid hard UI resets.
  }
}

async function sendGroupMessage() {
  const groupID = groupChatTargetID.value;
  const content = groupMessageDraft.value.trim();

  if (!groupID) {
    authError.value = "Choose a group chat first.";
    return;
  }
  if (!content) {
    authError.value = "Message content is required.";
    return;
  }

  loading.value = true;
  authError.value = "";

  try {
    const response = await fetch("/api/chat/groups/messages", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      credentials: "include",
      body: JSON.stringify({
        group_id: groupID,
        content,
      }),
    });

    const payload = await response.json();
    if (!response.ok) {
      authError.value = payload.error || "Unable to send group message.";
      return;
    }

    const sentMessage = payload.message;
    if (sentMessage && !groupMessages.value.some((message) => message.id === sentMessage.id)) {
      groupMessages.value = [...groupMessages.value, sentMessage];
    }
    groupMessageDraft.value = "";
  } catch (_err) {
    authError.value = "Group message sending failed.";
  } finally {
    loading.value = false;
  }
}

function closeRealtimeSocket() {
  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  if (!realtimeSocket) {
    return;
  }

  const socket = realtimeSocket;
  realtimeSocket = null;
  socket.onclose = null;
  socket.onerror = null;
  socket.onmessage = null;
  socket.close();
}

function connectRealtimeSocket() {
  if (!me.value) {
    return;
  }

  if (realtimeSocket && [WebSocket.OPEN, WebSocket.CONNECTING].includes(realtimeSocket.readyState)) {
    return;
  }

  if (reconnectTimer) {
    clearTimeout(reconnectTimer);
    reconnectTimer = null;
  }

  const protocol = window.location.protocol === "https:" ? "wss" : "ws";
  const socket = new WebSocket(`${protocol}://${window.location.host}/api/ws`);
  realtimeSocket = socket;

  socket.onmessage = (event) => {
    try {
      const payload = JSON.parse(event.data);
      handleRealtimePayload(payload);
    } catch (_err) {
      // Ignore malformed realtime payloads.
    }
  };

  socket.onerror = () => {
    socket.close();
  };

  socket.onclose = () => {
    if (realtimeSocket === socket) {
      realtimeSocket = null;
    }

    if (!me.value) {
      return;
    }

    reconnectTimer = setTimeout(() => {
      connectRealtimeSocket();
    }, 1500);
  };
}

function handleRealtimePayload(payload) {
  if (!payload || typeof payload !== "object") {
    return;
  }

  if (payload.type === "private_message") {
    maybeAppendPrivateMessage(payload.data);
    return;
  }

  if (payload.type === "group_message") {
    maybeAppendGroupMessage(payload.data);
    return;
  }

  if (payload.type === "notification") {
    prependNotification(payload.data);
    void refreshFromNotification(payload.data);
    return;
  }

  if (payload.type === "feed_updated") {
    void loadFeed();
    return;
  }

  if (payload.type === "groups_updated") {
    void Promise.allSettled([loadGroupsData(), loadSelectedGroupContent()]);
  }
}

async function refreshFromNotification(notification) {
  const type = notification?.type;

  if (type === "follow_request" || type === "follow_request_accepted" || type === "follow_request_declined") {
    await Promise.allSettled([loadNetworkData(), loadMyProfile()]);
    return;
  }

  if (type === "group_invite" || type === "group_join_request" || type === "group_event_created") {
    await loadGroupsData();
    return;
  }

  if (type === "private_message_received") {
    await Promise.allSettled([loadNotifications(), loadPrivateMessages()]);
    return;
  }

  if (type === "group_message_received") {
    await Promise.allSettled([loadNotifications(), loadGroupMessages()]);
  }
}

function maybeAppendPrivateMessage(message) {
  if (!message || !message.id || !me.value || !privateChatTargetID.value) {
    return;
  }

  const peerID = privateChatTargetID.value;
  const involvesPeer = message.sender_id === peerID || message.recipient_id === peerID;
  const involvesMe = message.sender_id === me.value.id || message.recipient_id === me.value.id;
  if (!involvesPeer || !involvesMe) {
    return;
  }

  if (privateMessages.value.some((existing) => existing.id === message.id)) {
    return;
  }

  privateMessages.value = [...privateMessages.value, message];
}

function maybeAppendGroupMessage(message) {
  if (!message || !message.id || !groupChatTargetID.value) {
    return;
  }

  if (message.group_id !== groupChatTargetID.value) {
    return;
  }

  if (groupMessages.value.some((existing) => existing.id === message.id)) {
    return;
  }

  groupMessages.value = [...groupMessages.value, message];
}

function prependNotification(notification) {
  if (!notification || !notification.id) {
    return;
  }

  const existingIndex = notifications.value.findIndex((existing) => existing.id === notification.id);
  if (existingIndex >= 0) {
    notifications.value[existingIndex] = {
      ...notifications.value[existingIndex],
      ...notification,
    };
    return;
  }

  notifications.value = [notification, ...notifications.value];
  if (!notification.is_read) {
    notificationsUnread.value += 1;
  }
}

function formatNotification(notification) {
  const payload = notification?.payload || {};

  if (notification?.type === "follow_request") {
    return `${payload.requester_name || "Someone"} sent you a follow request`;
  }
  if (notification?.type === "follow_request_accepted") {
    return `${payload.target_name || "Someone"} accepted your follow request`;
  }
  if (notification?.type === "follow_request_declined") {
    return `${payload.target_name || "Someone"} declined your follow request`;
  }
  if (notification?.type === "private_message_received") {
    return `${payload.sender_name || "Someone"} sent you a private message`;
  }
  if (notification?.type === "group_message_received") {
    return `${payload.sender_name || "Someone"} sent a message in ${payload.group_title || "your group"}`;
  }
  if (notification?.type === "group_invite") {
    return `${payload.inviter_name || "Someone"} invited you to ${payload.group_title || "a group"}`;
  }
  if (notification?.type === "group_join_request") {
    return `${payload.requester_name || "Someone"} requested to join ${payload.group_title || "your group"}`;
  }
  if (notification?.type === "group_event_created") {
    return `${payload.creator_name || "Someone"} created event ${payload.event_title || "in your group"}`;
  }

  return notification?.type || "notification";
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
    stopActivityPolling();
    closeRealtimeSocket();
    me.value = null;
    profile.value = null;
    networkUsers.value = [];
    incomingRequests.value = [];
    posts.value = [];
    groups.value = [];
    groupInvitesIncoming.value = [];
    groupJoinRequestsIncoming.value = [];
    selectedGroupID.value = "";
    groupPosts.value = [];
    groupEvents.value = [];
    notifications.value = [];
    notificationsUnread.value = 0;
    privateChatTargetID.value = "";
    privateMessages.value = [];
    privateMessageDraft.value = "";
    groupChatTargetID.value = "";
    groupMessages.value = [];
    groupMessageDraft.value = "";
    postForm.content = "";
    postForm.privacy = "public";
    postForm.media = null;
    postForm.allowed_user_ids = [];
    groupPostForm.content = "";
    groupPostForm.media = null;
    groupForm.title = "";
    groupForm.description = "";
    groupEventForm.title = "";
    groupEventForm.description = "";
    groupEventForm.event_datetime = "";
    groupEventForm.option_one = "Going";
    groupEventForm.option_two = "Not going";
    groupEventForm.option_three = "";
    profileViewUserID.value = "";
    profileViewResult.value = null;
    profileViewError.value = "";
    loading.value = false;
    loginForm.password = "";
  }
}
</script>
