import { useCallback, useEffect, useState } from 'react'
import { Link, useNavigate } from 'react-router-dom'
import { Repeat2 } from 'lucide-react'
import { apiRequest, apiUpload } from '../api/client'
import type { SocialPost } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { SkeletonFeed } from '../components/Skeleton'
import { SuggestionsPanel } from '../components/SuggestionsPanel'
import { SharePostModal } from '../components/SharePostModal'

function authorName(p: { author?: { first_name: string; last_name: string } }) {
  const a = p.author
  if (!a) return 'Membre'
  return `${a.first_name} ${a.last_name}`
}

function formatDate(iso: string) {
  return new Date(iso).toLocaleString('fr-FR', {
    day: 'numeric', month: 'short', hour: '2-digit', minute: '2-digit',
  })
}

function initials(name: string) {
  return name.split(' ').map((p) => p[0]).join('').slice(0, 2).toUpperCase()
}

export function FeedPage() {
  const navigate = useNavigate()
  const { token, user, requireAuth, openLogin } = useAuth()
  const [posts, setPosts] = useState<SocialPost[]>([])
  const [tab, setTab] = useState<'all' | 'following'>('all')
  const [error, setError] = useState('')
  const [ready, setReady] = useState(false)
  const [sharePost, setSharePost] = useState<SocialPost | null>(null)

  const [body, setBody] = useState('')
  const [clubId, setClubId] = useState('')
  const [files, setFiles] = useState<File[]>([])
  const [previews, setPreviews] = useState<string[]>([])
  const [publishing, setPublishing] = useState(false)
  const { clubs } = useAuth()

  const load = useCallback(async () => {
    setError('')
    try {
      const qs = tab === 'following' ? '?limit=15&following=1' : '?limit=15'
      const data = await apiRequest<{ posts: SocialPost[] }>(`/social/feed${qs}`, {}, token)
      setPosts(data.posts)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Impossible de charger le fil')
    } finally {
      setReady(true)
    }
  }, [token, tab])

  useEffect(() => { void load() }, [load])

  useEffect(() => {
    const urls = files.map((f) => URL.createObjectURL(f))
    setPreviews(urls)
    return () => urls.forEach((u) => URL.revokeObjectURL(u))
  }, [files])

  function onPickFiles(list: FileList | null) {
    if (!list) return
    const next = [...files, ...Array.from(list)].slice(0, 6)
    setFiles(next)
  }

  function publish() {
    requireAuth(async () => {
      setPublishing(true)
      setError('')
      try {
        const t = getToken()
        const form = new FormData()
        form.append('body', body.trim())
        if (clubId) form.append('club_id', clubId)
        files.forEach((f) => form.append('media', f))
        const created = await apiUpload<SocialPost>('/social/posts', form, t)
        setBody('')
        setFiles([])
        setClubId('')
        setPosts((prev) => [created, ...prev.filter((p) => p.id !== created.id)])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Publication impossible')
      } finally {
        setPublishing(false)
      }
    }, 'Connectez-vous pour publier')
  }

  function toggleLike(post: SocialPost) {
    requireAuth(async () => {
      const t = getToken()
      const nextLiked = !post.reacted_by_me
      const delta = nextLiked ? 1 : -1
      setPosts((prev) => prev.map((p) => (
        p.id === post.id
          ? {
              ...p,
              reacted_by_me: nextLiked,
              reaction_count: Math.max(0, p.reaction_count + delta),
            }
          : p
      )))
      try {
        if (post.reacted_by_me) {
          await apiRequest(`/social/posts/${post.id}/reactions`, { method: 'DELETE' }, t)
        } else {
          await apiRequest(`/social/posts/${post.id}/reactions`, {
            method: 'POST', body: JSON.stringify({ kind: 'like' }),
          }, t)
        }
      } catch {
        setPosts((prev) => prev.map((p) => (
          p.id === post.id
            ? {
                ...p,
                reacted_by_me: post.reacted_by_me,
                reaction_count: post.reaction_count,
              }
            : p
        )))
      }
    }, 'Connectez-vous pour aimer une publication')
  }

  function repost(post: SocialPost, quoteBody: string) {
    requireAuth(async () => {
      try {
        const created = await apiRequest<SocialPost>(
          `/social/posts/${post.id}/repost`,
          { method: 'POST', body: JSON.stringify({ quote_body: quoteBody }) },
          getToken(),
        )
        setPosts((prev) => [created, ...prev.filter((p) => p.id !== created.id)])
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Repartage impossible')
      }
    }, 'Connectez-vous pour repartager une publication')
  }

  function followAuthor(authorId: string) {
    requireAuth(async () => {
      setPosts((prev) => prev.map((p) => (
        p.author_id === authorId ? { ...p, author_followed_by_me: true } : p
      )))
      try {
        await apiRequest(`/social/users/${authorId}/follow`, { method: 'POST', body: '{}' }, getToken())
      } catch {
        setPosts((prev) => prev.map((p) => (
          p.author_id === authorId ? { ...p, author_followed_by_me: false } : p
        )))
      }
    }, 'Connectez-vous pour vous abonner')
  }

  async function deletePost(postId: string) {
    if (!confirm('Supprimer cette publication ?')) return
    const snapshot = posts
    setPosts((prev) => prev.filter((p) => p.id !== postId))
    try {
      await apiRequest(`/social/posts/${postId}`, { method: 'DELETE' }, getToken())
    } catch {
      setPosts(snapshot)
    }
  }

  function openPost(postId: string) {
    navigate(`/posts/${postId}`)
  }

  function openComments(postId: string) {
    navigate(`/posts/${postId}?focus=comments`)
  }

  const showSkeleton = !ready && posts.length === 0
  const showEmpty = ready && posts.length === 0

  return (
    <div className="stack gap-md">
      {!user && (
        <div className="guest-banner">
          <div>
            <strong>Bienvenue sur Rotaract Social</strong>
            <p>Parcourez le fil librement. Connectez-vous pour aimer, commenter, publier ou vous abonner.</p>
          </div>
          <button type="button" className="btn-primary" onClick={() => openLogin('Rejoignez la conversation')}>
            Se connecter
          </button>
        </div>
      )}

      <div className="feed-tabs">
        <button type="button" className={tab === 'all' ? 'active' : ''} onClick={() => setTab('all')}>Pour vous</button>
        <button
          type="button"
          className={tab === 'following' ? 'active' : ''}
          onClick={() => {
            if (!user) {
              openLogin('Connectez-vous pour voir vos abonnements')
              return
            }
            setTab('following')
          }}
        >
          Abonnements
        </button>
      </div>

      <section className="composer-card">
        <div className="composer-row">
          <div className="avatar-circle lg">
            {user?.avatar_url
              ? <img src={user.avatar_url} alt="" />
              : initials(user ? `${user.first_name} ${user.last_name}` : 'R')}
          </div>
          <button
            type="button"
            className="composer-fake"
            onClick={() => {
              if (!user) openLogin('Connectez-vous pour publier')
              else document.getElementById('composer-input')?.focus()
            }}
          >
            {user ? `Quoi de neuf, ${user.first_name} ?` : 'Quoi de neuf sur Rotaract ?'}
          </button>
        </div>
        <textarea
          id="composer-input"
          value={body}
          onChange={(e) => setBody(e.target.value)}
          placeholder="Partagez une actualité, une photo ou une vidéo…"
          rows={3}
        />
        {previews.length > 0 && (
          <div className="media-preview">
            {files.map((f, i) => (
              f.type.startsWith('video/')
                ? <video key={previews[i]} src={previews[i]} controls />
                : <img key={previews[i]} src={previews[i]} alt="" />
            ))}
          </div>
        )}
        <div className="composer-actions">
          <label className="media-btn">
            Photo / Vidéo
            <input
              type="file"
              accept="image/jpeg,image/png,image/webp,video/mp4,video/webm"
              multiple
              hidden
              onChange={(e) => {
                if (!user) {
                  openLogin('Connectez-vous pour ajouter des médias')
                  e.target.value = ''
                  return
                }
                onPickFiles(e.target.files)
              }}
            />
          </label>
          {clubs.length > 0 && (
            <select value={clubId} onChange={(e) => setClubId(e.target.value)} aria-label="Club">
              <option value="">Sans club</option>
              {clubs.map((m) => (
                <option key={m.club_id} value={m.club_id}>{m.club?.name ?? m.club_id}</option>
              ))}
            </select>
          )}
          <button
            type="button"
            className="btn-primary"
            disabled={publishing || (!body.trim() && files.length === 0)}
            onClick={publish}
          >
            {publishing ? 'Publication…' : 'Publier'}
          </button>
        </div>
      </section>

      {error && <p className="error" role="alert">{error}</p>}

      {showSkeleton && <SkeletonFeed count={3} />}

      {showEmpty && (
        <div className="empty">
          <strong>Le fil est calme</strong>
          <p>Soyez le premier à publier une photo, une vidéo ou une actualité.</p>
        </div>
      )}

      <div className="feed">
        {posts.map((post) => (
          <article key={post.id} className="post-card">
            <header className="post-head">
              <div className="post-author">
                <Link to={`/users/${post.author_id}`} className="avatar-circle">
                  {post.author?.avatar_url
                    ? <img src={post.author.avatar_url} alt="" />
                    : initials(authorName(post))}
                </Link>
                <div>
                  <strong>
                    <Link to={`/users/${post.author_id}`}>{authorName(post)}</Link>
                  </strong>
                  {post.club && <span className="club-tag">{post.club.name}</span>}
                  <p className="muted small">{formatDate(post.created_at)}</p>
                </div>
              </div>
              <div className="post-head-actions">
                {user && user.id !== post.author_id && !post.author_followed_by_me && (
                  <button type="button" className="btn-tiny btn-secondary" onClick={() => followAuthor(post.author_id)}>
                    Suivre
                  </button>
                )}
                {(user?.id === post.author_id || user?.is_admin) && (
                  <button type="button" className="btn-ghost small" onClick={() => deletePost(post.id)}>Supprimer</button>
                )}
              </div>
            </header>

            {post.reposted_post_id && <p className="repost-label"><Repeat2 size={14} aria-hidden="true" /> Repartagé</p>}
            {(post.quote_body || post.body) && (
              <button type="button" className="post-body post-body-btn" onClick={() => openPost(post.id)}>
                {post.quote_body || post.body}
              </button>
            )}

            {post.original && (
              <button
                type="button"
                className="original-post-preview"
                onClick={() => openPost(post.original!.id)}
              >
                <p className="repost-label">Publication originale de {authorName(post.original)}</p>
                {post.original.body && <p className="post-body">{post.original.body}</p>}
                {post.original.media?.[0] && (
                  post.original.media[0].kind === 'video'
                    ? <video src={post.original.media[0].url} muted playsInline />
                    : <img src={post.original.media[0].url} alt="" loading="lazy" />
                )}
              </button>
            )}

            {post.media && post.media.length > 0 && (
              <button
                type="button"
                className={`media-grid count-${Math.min(post.media.length, 4)} media-grid-btn`}
                onClick={() => openPost(post.id)}
              >
                {post.media.slice(0, 4).map((m, index) => (
                  <span className="media-tile" key={m.id}>
                    {m.kind === 'video'
                      ? <video src={m.url} muted playsInline />
                      : <img src={m.url} alt="" loading="lazy" />}
                    {index === 3 && post.media!.length > 4 && (
                      <span className="media-more">+{post.media!.length - 4}</span>
                    )}
                  </span>
                ))}
              </button>
            )}

            <div className="post-stats muted small">
              <span>{post.reaction_count} j&apos;aime</span>
              <button
                type="button"
                className="post-stats-link"
                onClick={() => openComments(post.id)}
              >
                {post.comment_count} commentaires
              </button>
            </div>

            <footer className="post-actions">
              <button type="button" className={post.reacted_by_me ? 'action active' : 'action'} onClick={() => toggleLike(post)}>
                J&apos;aime
              </button>
              <button type="button" className="action" onClick={() => openComments(post.id)}>
                Commenter
              </button>
              <button
                type="button"
                className="action"
                onClick={() => setSharePost(post)}
              >
                Repartager
              </button>
            </footer>
          </article>
        ))}
      </div>

      <SuggestionsPanel />
      <SharePostModal
        post={sharePost}
        onClose={() => setSharePost(null)}
        onRepost={(quoteBody) => {
          if (sharePost) repost(sharePost, quoteBody)
        }}
      />
    </div>
  )
}
