import { useCallback, useEffect, useRef, useState, type FormEvent } from 'react'
import { Link, useNavigate, useParams, useSearchParams } from 'react-router-dom'
import { ChevronLeft, ChevronRight, Repeat2 } from 'lucide-react'
import { apiRequest } from '../api/client'
import type { SocialComment, SocialPost } from '../api/types'
import { useAuth } from '../auth/AuthContext'
import { getToken } from '../auth/storage'
import { ShareCommentModal } from '../components/ShareCommentModal'
import { SharePostModal } from '../components/SharePostModal'
import { SkeletonFeed, SkeletonList } from '../components/Skeleton'

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

export function PostDetailPage() {
  const { postID } = useParams()
  const [searchParams] = useSearchParams()
  const focusComments = searchParams.get('focus') === 'comments'
  const navigate = useNavigate()
  const { token, user, requireAuth, openLogin } = useAuth()
  const [post, setPost] = useState<SocialPost | null>(null)
  const [comments, setComments] = useState<SocialComment[]>([])
  const [commentBody, setCommentBody] = useState('')
  const [replyTo, setReplyTo] = useState<string | null>(null)
  const [replyName, setReplyName] = useState('')
  const [shareComment, setShareComment] = useState<SocialComment | null>(null)
  const [sharePost, setSharePost] = useState<SocialPost | null>(null)
  const [error, setError] = useState('')
  const [postReady, setPostReady] = useState(false)
  const [commentsReady, setCommentsReady] = useState(false)
  const [sending, setSending] = useState(false)
  const [activeMediaIndex, setActiveMediaIndex] = useState(0)
  const inputRef = useRef<HTMLInputElement>(null)
  const rightRef = useRef<HTMLElement>(null)
  const focusedOnce = useRef(false)
  const lastPostID = useRef<string | undefined>(undefined)
  const touchStartX = useRef<number | null>(null)

  // Reset only when navigating to a different post
  useEffect(() => {
    if (postID !== lastPostID.current) {
      lastPostID.current = postID
      setPost(null)
      setComments([])
      setPostReady(false)
      setCommentsReady(false)
      setError('')
      setActiveMediaIndex(0)
      focusedOnce.current = false
    }
  }, [postID])

  const loadPost = useCallback(async () => {
    if (!postID) return
    setError('')
    try {
      const data = await apiRequest<SocialPost>(`/social/posts/${postID}`, {}, token)
      setPost(data)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Publication introuvable')
      setPost(null)
    } finally {
      setPostReady(true)
    }
  }, [postID, token])

  const loadComments = useCallback(async () => {
    if (!postID) return
    try {
      const data = await apiRequest<{ comments: SocialComment[] }>(
        `/social/posts/${postID}/comments`,
        {},
        token,
      )
      setComments(data.comments)
    } catch {
      /* keep existing thread on soft refresh failure */
    } finally {
      setCommentsReady(true)
    }
  }, [postID, token])

  useEffect(() => { void loadPost() }, [loadPost])
  useEffect(() => { void loadComments() }, [loadComments])

  useEffect(() => {
    if (!focusComments || !postReady || !post || focusedOnce.current) return
    focusedOnce.current = true
    rightRef.current?.scrollIntoView({ behavior: 'smooth', block: 'start' })
    window.setTimeout(() => {
      inputRef.current?.focus()
      if (!user) openLogin('Connectez-vous pour commenter')
    }, 120)
  }, [focusComments, postReady, post, user, openLogin])

  function toggleLike() {
    if (!post) return
    requireAuth(async () => {
      const t = getToken()
      const prev = post
      const nextLiked = !post.reacted_by_me
      setPost({
        ...post,
        reacted_by_me: nextLiked,
        reaction_count: Math.max(0, post.reaction_count + (nextLiked ? 1 : -1)),
      })
      try {
        if (prev.reacted_by_me) {
          await apiRequest(`/social/posts/${post.id}/reactions`, { method: 'DELETE' }, t)
        } else {
          await apiRequest(`/social/posts/${post.id}/reactions`, {
            method: 'POST', body: JSON.stringify({ kind: 'like' }),
          }, t)
        }
      } catch {
        setPost(prev)
      }
    }, 'Connectez-vous pour aimer une publication')
  }

  function repost(quoteBody: string) {
    if (!post) return
    requireAuth(async () => {
      try {
        await apiRequest(`/social/posts/${post.id}/repost`, {
          method: 'POST',
          body: JSON.stringify({ quote_body: quoteBody }),
        }, getToken())
        alert('Publication repartagée')
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Repartage impossible')
      }
    }, 'Connectez-vous pour repartager une publication')
  }

  function startReply(c: SocialComment) {
    requireAuth(() => {
      setReplyTo(c.id)
      setReplyName(authorName(c))
      inputRef.current?.focus()
    }, 'Connectez-vous pour répondre')
  }

  function submitComment(e: FormEvent) {
    e.preventDefault()
    if (!postID || sending) return
    const body = commentBody.trim()
    if (!body) return
    requireAuth(async () => {
      const t = getToken()
      const me = user
      if (!me) return
      const parentId = replyTo
      const parentLabel = replyName
      const optimistic: SocialComment = {
        id: `tmp-${crypto.randomUUID()}`,
        post_id: postID,
        author_id: me.id,
        parent_id: parentId || undefined,
        body,
        created_at: new Date().toISOString(),
        author: {
          id: me.id,
          first_name: me.first_name,
          last_name: me.last_name,
          avatar_url: me.avatar_url,
        },
        replies: [],
      }

      setCommentBody('')
      setReplyTo(null)
      setReplyName('')
      setSending(true)

      if (parentId) {
        setComments((prev) => prev.map((c) => (
          c.id === parentId
            ? { ...c, replies: [...(c.replies ?? []), optimistic] }
            : c
        )))
      } else {
        setComments((prev) => [...prev, optimistic])
      }
      setPost((prev) => prev ? { ...prev, comment_count: prev.comment_count + 1 } : prev)

      try {
        const created = await apiRequest<SocialComment>(`/social/posts/${postID}/comments`, {
          method: 'POST',
          body: JSON.stringify({ body, parent_id: parentId || undefined }),
        }, t)

        if (parentId) {
          setComments((prev) => prev.map((c) => (
            c.id === parentId
              ? {
                  ...c,
                  replies: (c.replies ?? []).map((r) => (
                    r.id === optimistic.id ? { ...created, replies: [] } : r
                  )),
                }
              : c
          )))
        } else {
          setComments((prev) => prev.map((c) => (
            c.id === optimistic.id ? { ...created, replies: [] } : c
          )))
        }
      } catch {
        if (parentId) {
          setComments((prev) => prev.map((c) => (
            c.id === parentId
              ? { ...c, replies: (c.replies ?? []).filter((r) => r.id !== optimistic.id) }
              : c
          )))
        } else {
          setComments((prev) => prev.filter((c) => c.id !== optimistic.id))
        }
        setPost((prev) => prev ? { ...prev, comment_count: Math.max(0, prev.comment_count - 1) } : prev)
        setCommentBody(body)
        if (parentId) {
          setReplyTo(parentId)
          setReplyName(parentLabel)
        }
      } finally {
        setSending(false)
      }
    }, 'Connectez-vous pour commenter')
  }

  function renderComment(c: SocialComment, nested = false) {
    return (
      <div key={c.id} className={nested ? 'thread-item nested' : 'thread-item'}>
        <Link to={`/users/${c.author_id}`} className="avatar-circle xs">
          {c.author?.avatar_url ? <img src={c.author.avatar_url} alt="" /> : initials(authorName(c))}
        </Link>
        <div className="thread-body">
          <p className="thread-text">
            <Link to={`/users/${c.author_id}`} className="thread-name">{authorName(c)}</Link>
            {' '}
            {c.body}
          </p>
          <div className="thread-meta">
            <time className="muted small">{formatDate(c.created_at)}</time>
            {!nested && (
              <button type="button" className="thread-action" onClick={() => startReply(c)}>
                Répondre
              </button>
            )}
            <button
              type="button"
              className="thread-action"
              onClick={() => requireAuth(() => setShareComment(c), 'Connectez-vous pour partager')}
            >
              Envoyer
            </button>
          </div>
          {c.replies && c.replies.length > 0 && (
            <div className="thread-replies">
              {c.replies.map((r) => renderComment(r, true))}
            </div>
          )}
        </div>
      </div>
    )
  }

  if (!postReady && !post) {
    return (
      <div className="post-detail">
        <div className="post-detail-stage"><SkeletonFeed count={1} /></div>
        <div className="post-detail-right"><SkeletonList count={4} /></div>
      </div>
    )
  }

  if (error || !post) {
    return (
      <div className="stack gap-md" style={{ padding: '1.25rem' }}>
        <button type="button" className="btn-text" onClick={() => navigate(-1)}>← Retour</button>
        <p className="error">{error || 'Publication introuvable'}</p>
      </div>
    )
  }

  const displayPost = post.original ?? post
  const hasMedia = Boolean(displayPost.media && displayPost.media.length > 0)
  const media = displayPost.media ?? []
  const showCommentsSkeleton = !commentsReady && comments.length === 0

  function moveMedia(direction: -1 | 1) {
    setActiveMediaIndex((current) => (current + direction + media.length) % media.length)
  }

  function handleMediaPointerDown(event: React.PointerEvent<HTMLDivElement>) {
    if (event.pointerType === 'mouse' && event.button !== 0) return
    touchStartX.current = event.clientX
  }

  function handleMediaPointerUp(event: React.PointerEvent<HTMLDivElement>) {
    if (touchStartX.current === null || media.length < 2) return
    const distance = event.clientX - touchStartX.current
    touchStartX.current = null
    if (Math.abs(distance) < 48) return
    moveMedia(distance < 0 ? 1 : -1)
  }

  return (
    <>
      <div className={`post-detail${hasMedia ? ' has-media' : ''}${focusComments ? ' focus-comments' : ''}`}>
        <section className="post-detail-left">
          <button type="button" className="btn-text post-detail-back" onClick={() => navigate(-1)}>
            ← Retour
          </button>

          <div className="post-detail-stage">
            {hasMedia ? (
              <div
                className="post-detail-media"
                onPointerDown={handleMediaPointerDown}
                onPointerUp={handleMediaPointerUp}
              >
                <div className="post-detail-slide">
                  {media[activeMediaIndex].kind === 'video'
                    ? <video src={media[activeMediaIndex].url} controls playsInline />
                    : <img src={media[activeMediaIndex].url} alt={`Image ${activeMediaIndex + 1} sur ${media.length}`} />}
                </div>
                {media.length > 1 && (
                  <>
                    <button
                      type="button"
                      className="media-nav media-nav-prev"
                      onClick={() => moveMedia(-1)}
                      aria-label="Média précédent"
                    >
                      <ChevronLeft size={20} aria-hidden="true" />
                    </button>
                    <button
                      type="button"
                      className="media-nav media-nav-next"
                      onClick={() => moveMedia(1)}
                      aria-label="Média suivant"
                    >
                      <ChevronRight size={20} aria-hidden="true" />
                    </button>
                    <div className="media-dots" aria-label="Sélectionner un média">
                      {media.map((item, index) => (
                        <button
                          key={item.id}
                          type="button"
                          className={index === activeMediaIndex ? 'active' : ''}
                          onClick={() => setActiveMediaIndex(index)}
                          aria-label={`Afficher le média ${index + 1}`}
                          aria-current={index === activeMediaIndex ? 'true' : undefined}
                        />
                      ))}
                    </div>
                  </>
                )}
              </div>
            ) : (
              <div className="post-detail-text-only">
                {post.reposted_post_id && <p className="repost-label"><Repeat2 size={14} aria-hidden="true" /> Repartagé</p>}
                {post.quote_body && <p className="post-body xl">{post.quote_body}</p>}
                {post.original && <p className="post-body xl original-detail-text">{post.original.body}</p>}
                {!post.original && <p className="post-body xl">{post.body}</p>}
              </div>
            )}
          </div>

          <div className="post-detail-bar">
            <div className="post-stats muted small">
              <span>{post.reaction_count} j&apos;aime</span>
              <span>{post.comment_count} commentaires</span>
            </div>
            <footer className="post-actions">
              <button
                type="button"
                className={post.reacted_by_me ? 'action active' : 'action'}
                onClick={toggleLike}
              >
                J&apos;aime
              </button>
              <button type="button" className="action" onClick={() => setSharePost(post)}>
                Repartager
              </button>
            </footer>
          </div>
        </section>

        <aside
          ref={rightRef}
          className="post-detail-right"
          id="post-comments"
        >
          <header className="post-detail-author">
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
          </header>

          {(post.quote_body || (!post.original && post.body)) && (
            <p className="post-detail-caption">{post.quote_body || post.body}</p>
          )}
          {post.reposted_post_id && <p className="repost-label"><Repeat2 size={14} aria-hidden="true" /> Repartagé avec la publication originale ci-dessus</p>}

          <div className="post-detail-comments-head">
            <h2>Discussion</h2>
            <p className="muted small">{post.comment_count}</p>
          </div>

          <div className="post-detail-comments-scroll">
            {showCommentsSkeleton && <SkeletonList count={3} />}
            {commentsReady && comments.length === 0 && (
              <p className="muted thread-empty">Aucun commentaire. Lancez la discussion.</p>
            )}
            {comments.length > 0 && (
              <div className="thread">
                {comments.map((c) => renderComment(c))}
              </div>
            )}
          </div>

          <div className="post-detail-composer">
            {replyTo && (
              <p className="thread-replying">
                Réponse à <strong>{replyName}</strong>
                <button type="button" className="btn-ghost" onClick={() => { setReplyTo(null); setReplyName('') }}>
                  Annuler
                </button>
              </p>
            )}
            <form onSubmit={submitComment} className="thread-composer">
              <input
                ref={inputRef}
                id="post-comment-input"
                value={commentBody}
                onChange={(e) => setCommentBody(e.target.value)}
                placeholder={
                  !user
                    ? 'Connectez-vous pour commenter…'
                    : replyTo
                      ? `Répondre à ${replyName}…`
                      : 'Ajouter un commentaire…'
                }
                onFocus={() => { if (!user) openLogin('Connectez-vous pour commenter') }}
              />
              <button type="submit" className="btn-primary btn-tiny" disabled={!commentBody.trim() || sending}>
                Publier
              </button>
            </form>
          </div>
        </aside>
      </div>

      <ShareCommentModal comment={shareComment} onClose={() => setShareComment(null)} />
      <SharePostModal
        post={sharePost}
        onClose={() => setSharePost(null)}
        onRepost={(quoteBody) => repost(quoteBody)}
      />
    </>
  )
}
