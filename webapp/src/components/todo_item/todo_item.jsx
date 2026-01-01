import React, { useState, useRef, useCallback } from 'react';
import PropTypes from 'prop-types';

import { changeOpacity, makeStyleFromTheme } from 'mattermost-redux/utils/theme_utils';
import TextareaAutosize from 'react-textarea-autosize';

import CompleteButton from '../buttons/complete';
import AcceptButton from '../buttons/accept';
import {
    canComplete,
    canRemove,
    canAccept,
    canBump,
    handleFormattedTextClick,
} from '../../utils';
import CompassIcon from '../icons/compassIcons';
import Menu from '../../widget/menu';
import MenuItem from '../../widget/menuItem';
import MenuWrapper from '../../widget/menuWrapper';
import Button from '../../widget/buttons/button';

import PostPermalink from './post_permalink';

const PostUtils = window.PostUtils; // import the post utilities

function TodoItem(props) {
    const { issue, theme, siteURL, accept, complete, list, remove, bump, openTodoToast, openAssigneeModal, setEditingTodo, editIssue, fetchComments, addComment, deleteComment, currentUserId } = props;
    const [done, setDone] = useState(false);
    const [editTodo, setEditTodo] = useState(false);
    const [message, setMessage] = useState(issue.message);
    const [description, setDescription] = useState(issue.description);
    const MONTHS = ['Jan', 'Feb', 'Mar', 'Apr', 'May', 'Jun', 'Jul', 'Aug', 'Sep', 'Oct', 'Nov', 'Dec'];
    const [hidden, setHidden] = useState(false);
    const [priority, setPriority] = useState(issue.priority || 0);
    const [dueAt, setDueAt] = useState(issue.due_at ? new Date(issue.due_at).toISOString().split('T')[0] : '');

    const date = new Date(issue.create_at);
    const year = date.getFullYear();
    const month = MONTHS[date.getMonth()];
    const day = date.getDate();
    const hours = date.getHours();
    const minutes = '0' + date.getMinutes();
    const seconds = '0' + date.getSeconds();
    const formattedTime = hours + ':' + minutes.substr(-2) + ':' + seconds.substr(-2);
    const formattedDate = month + ' ' + day + ', ' + year;

    const [showComments, setShowComments] = useState(false);
    const [comments, setComments] = useState([]);
    const [newComment, setNewComment] = useState('');
    const [loadingComments, setLoadingComments] = useState(false);

    React.useEffect(() => {
        if (showComments) {
            setLoadingComments(true);
            fetchComments(issue.id).then((data) => {
                setLoadingComments(false);
                if (!data.error) {
                    setComments(data);
                }
            });
        }
    }, [showComments, issue.id, fetchComments]);

    const handleAddComment = async () => {
        if (!newComment.trim()) {
            return;
        }
        const result = await addComment(issue.id, newComment);
        if (!result.error) {
            setNewComment('');
            const updatedComments = await fetchComments(issue.id);
            if (!updatedComments.error) {
                setComments(updatedComments);
            }
        }
    };

    const handleDeleteComment = async (commentId) => {
        const result = await deleteComment(commentId);
        if (!result.error) {
            const updatedComments = await fetchComments(issue.id);
            if (!updatedComments.error) {
                setComments(updatedComments);
            }
        }
    };

    const style = getStyle(theme);

    const handleClick = (e) => handleFormattedTextClick(e);

    const htmlFormattedMessage = PostUtils.formatText(issue.message, {
        siteURL,
    });

    const htmlFormattedDescription = PostUtils.formatText(issue.description, {
        siteURL,
    });

    const issueMessage = PostUtils.messageHtmlToComponent(htmlFormattedMessage);
    const issueDescription = PostUtils.messageHtmlToComponent(htmlFormattedDescription);

    let listPositionMessage = '';
    let createdMessage = 'Created ';
    if (issue.user) {
        if (issue.list === '') {
            createdMessage = 'Sent to ' + issue.user;
            listPositionMessage =
                'Accepted. On position ' + (issue.position + 1) + '.';
        } else if (issue.list === 'in') {
            createdMessage = 'Sent to ' + issue.user;
            listPositionMessage =
                'In Inbox on position ' + (issue.position + 1) + '.';
        } else if (issue.list === 'out') {
            createdMessage = 'Received from ' + issue.user;
            listPositionMessage = '';
        }
    }

    const listDiv = (
        <div
            className='light'
            style={style.subtitle}
        >
            {listPositionMessage}
        </div>
    );

    const acceptButton = (
        <AcceptButton
            issueId={issue.id}
            accept={accept}
        />
    );

    const onKeyDown = (e) => {
        if (e.key === 'Enter' && !e.shiftKey) {
            saveEditedTodo();
        }

        if (e.key === 'Escape') {
            setEditTodo(false);
        }
    };

    const actionButtons = (
        <div className='todo-action-buttons'>
            {canAccept(list) && acceptButton}
        </div>
    );

    const completeTimeout = useRef(null);
    const removeTimeout = useRef(null);

    const completeToast = useCallback(() => {
        openTodoToast({ icon: 'check', message: 'Todo completed', undo: undoCompleteTodo });

        setHidden(true);

        completeTimeout.current = setTimeout(() => {
            complete(issue.id);
        }, 5000);
    }, [complete, openTodoToast, issue]);

    const undoRemoveTodo = () => {
        clearTimeout(removeTimeout.current);
        setHidden(false);
    };

    const undoCompleteTodo = () => {
        clearTimeout(completeTimeout.current);
        setHidden(false);
        setDone(false);
    };

    const completeButton = (
        <CompleteButton
            active={done}
            theme={theme}
            issueId={issue.id}
            markAsDone={() => setDone(true)}
            completeToast={completeToast}
        />
    );

    const removeTodo = useCallback(() => {
        openTodoToast({ icon: 'trash-can-outline', message: 'Todo deleted', undo: undoRemoveTodo });
        setHidden(true);
        removeTimeout.current = setTimeout(() => {
            remove(issue.id);
        }, 5000);
    }, [remove, issue.id, openTodoToast]);

    const saveEditedTodo = () => {
        setEditTodo(false);
        const dueAtTimestamp = dueAt ? new Date(dueAt).getTime() : 0;
        editIssue(issue.id, message, description, dueAtTimestamp, priority);
    };

    const editAssignee = () => {
        openAssigneeModal('');
        setEditingTodo(issue.id);
    };

    return (
        <div
            key={issue.id}
            className={`todo-item ${done ? 'todo-item--done' : ''} ${hidden ? 'todo-item--hidden' : ''} `}
        >
            <div style={style.todoTopContent}>
                <div className='todo-item__content'>
                    {(canComplete(list)) && completeButton}
                    <div style={style.itemContent}>
                        {editTodo && (
                            <div>
                                <TextareaAutosize
                                    style={style.textareaResizeMessage}
                                    placeholder='Enter a title'
                                    value={message}
                                    autoFocus={true}
                                    onKeyDown={(e) => onKeyDown(e)}
                                    onChange={(e) => setMessage(e.target.value)}
                                />
                                <TextareaAutosize
                                    style={style.textareaResizeDescription}
                                    placeholder='Enter a description'
                                    value={description}
                                    onKeyDown={(e) => onKeyDown(e)}
                                    onChange={(e) => setDescription(e.target.value)}
                                />
                                <div style={style.enterpriseOptions}>
                                    <div style={style.optionItem}>
                                        <label style={style.optionLabel}>{'Priority'}</label>
                                        <select
                                            style={style.select}
                                            value={priority}
                                            onChange={(e) => setPriority(parseInt(e.target.value, 10))}
                                        >
                                            <option value={0}>{'Low'}</option>
                                            <option value={1}>{'Medium'}</option>
                                            <option value={2}>{'High'}</option>
                                        </select>
                                    </div>
                                    <div style={style.optionItem}>
                                        <label style={style.optionLabel}>{'Due Date'}</label>
                                        <input
                                            type='date'
                                            style={style.dateInput}
                                            value={dueAt}
                                            onChange={(e) => setDueAt(e.target.value)}
                                        />
                                    </div>
                                </div>
                            </div>
                        )}

                        {!editTodo && (
                            <div
                                className='todo-text'
                                onClick={handleClick}
                            >
                                {issueMessage}
                                {issue.postPermalink && <PostPermalink postPermalink={issue.postPermalink} />}

                                <div style={style.badgeContainer}>
                                    {issue.priority === 2 && <span style={style.priorityHigh}>{'High'}</span>}
                                    {issue.priority === 1 && <span style={style.priorityMedium}>{'Medium'}</span>}
                                    {issue.due_at > 0 && (
                                        <span style={style.dueDate}>
                                            <CompassIcon icon='calendar-outline' style={{ fontSize: 12, marginRight: 4 }} />
                                            {new Date(issue.due_at).toLocaleDateString()}
                                        </span>
                                    )}
                                </div>

                                <div style={style.description}>{issueDescription}</div>
                                {(canRemove(list, issue.list) ||
                                    canComplete(list) ||
                                    canAccept(list)) &&
                                    actionButtons}
                                {issue.user && (
                                    <div
                                        className='light'
                                        style={style.subtitle}
                                    >
                                        {createdMessage + ' on ' + formattedDate + ' at ' + formattedTime}
                                    </div>
                                )}
                                <div
                                    style={style.commentToggle}
                                    onClick={(e) => {
                                        e.stopPropagation();
                                        setShowComments(!showComments);
                                    }}
                                >
                                    <CompassIcon
                                        icon='comment-outline'
                                        style={{ fontSize: 14, marginRight: 4 }}
                                    />
                                    {showComments ? 'Hide Comments' : 'Comments'}
                                </div>
                                {listPositionMessage && listDiv}
                            </div>
                        )}
                    </div>
                </div>

                {showComments && (
                    <div style={style.commentsSection}>
                        <div style={style.commentsList}>
                            {loadingComments && <div style={style.noComments}>{'Loading comments...'}</div>}
                            {!loadingComments && comments.map((c) => (
                                <div key={c.id} style={style.commentItem} className='todo-comment-item'>
                                    <div style={style.commentHeader}>
                                        <span style={style.commentUser}>{c.username}</span>
                                        <div style={{ display: 'flex', alignItems: 'center' }}>
                                            <span style={style.commentDate}>{new Date(c.created_at).toLocaleString()}</span>
                                            {c.user_id === currentUserId && (
                                                <CompassIcon
                                                    icon='delete-outline'
                                                    style={style.commentDeleteIcon}
                                                    onClick={() => handleDeleteComment(c.id)}
                                                    className='todo-comment-delete'
                                                />
                                            )}
                                        </div>
                                    </div>
                                    <div style={style.commentMessage}>{c.message}</div>
                                </div>
                            ))}
                            {comments.length === 0 && <div style={style.noComments}>{'No comments yet.'}</div>}
                        </div>
                        <div style={style.addCommentContainer}>
                            <TextareaAutosize
                                style={style.commentInput}
                                placeholder='Add a comment…'
                                value={newComment}
                                onChange={(e) => setNewComment(e.target.value)}
                            />
                            <Button
                                emphasis='primary'
                                size='xsmall'
                                onClick={handleAddComment}
                                disabled={!newComment.trim()}
                                style={{ marginTop: 4 }}
                            >
                                {'Post'}
                            </Button>
                        </div>
                    </div>
                )}
                {!editTodo && (
                    <MenuWrapper>
                        <button className='todo-item__dots'>
                            <CompassIcon icon='dots-vertical' />
                        </button>
                        <Menu position='left'>
                            {canAccept(list) && (
                                <MenuItem
                                    action={() => accept(issue.id)}
                                    text='Accept todo'
                                    icon='check'
                                />
                            )}
                            {canBump(list, issue.list) && (
                                <MenuItem
                                    text='Bump'
                                    icon='bell-outline'
                                    action={() => bump(issue.id)}
                                />
                            )}
                            <MenuItem
                                text='Edit todo'
                                icon='pencil-outline'
                                action={() => setEditTodo(true)}
                                shortcut='e'
                            />
                            <MenuItem
                                text='Assign to…'
                                icon='account-plus-outline'
                                action={editAssignee}
                                shortcut='a'
                            />
                            {canRemove(list, issue.list) && (
                                <MenuItem
                                    action={removeTodo}
                                    text='Delete todo'
                                    icon='trash-can-outline'
                                    shortcut='d'
                                />
                            )}
                        </Menu>
                    </MenuWrapper>
                )}
            </div>
            {editTodo &&
                (
                    <div
                        className='todoplugin-button-container'
                        style={style.buttons}
                    >
                        <Button
                            emphasis='tertiary'
                            size='small'
                            onClick={() => setEditTodo(false)}
                        >
                            {'Cancel'}
                        </Button>
                        <Button
                            emphasis='primary'
                            size='small'
                            onClick={saveEditedTodo}
                        >
                            {'Save'}
                        </Button>
                    </div>
                )}
        </div>
    );
}

const getStyle = makeStyleFromTheme((theme) => {
    return {
        container: {
            padding: '8px 20px',
            display: 'flex',
            alignItems: 'flex-start',
        },
        itemContent: {
            width: '100%',
            display: 'flex',
            alignItems: 'center',
        },
        todoTopContent: {
            display: 'flex',
            justifyContent: 'space-between',
            flex: 1,
        },
        issueTitle: {
            color: theme.centerChannelColor,
            lineHeight: 1.7,
            fontWeight: 'bold',
        },
        subtitle: {
            marginTop: '4px',
            fontStyle: 'italic',
            fontSize: '13px',
        },
        message: {
            width: '100%',
            overflowWrap: 'break-word',
            whiteSpace: 'pre-wrap',
        },
        description: {
            marginTop: 4,
            fontSize: 12,
            color: changeOpacity(theme.centerChannelColor, 0.72),
        },
        buttons: {
            padding: '10px 0',
        },
        textareaResizeMessage: {
            border: 0,
            padding: 0,
            fontSize: 14,
            width: '100%',
            backgroundColor: 'transparent',
            resize: 'none',
            boxShadow: 'none',
        },
        textareaResizeDescription: {
            fontSize: 12,
            color: changeOpacity(theme.centerChannelColor, 0.72),
            marginTop: 1,
            border: 0,
            padding: 0,
            width: '100%',
            backgroundColor: 'transparent',
            resize: 'none',
            boxShadow: 'none',
        },
        enterpriseOptions: {
            display: 'flex',
            marginTop: 12,
            gap: 16,
        },
        optionItem: {
            display: 'flex',
            flexDirection: 'column',
            flex: 1,
        },
        optionLabel: {
            fontSize: 11,
            fontWeight: 600,
            marginBottom: 4,
            color: changeOpacity(theme.centerChannelColor, 0.64),
        },
        select: {
            height: 28,
            borderRadius: 4,
            border: `1px solid ${changeOpacity(theme.centerChannelColor, 0.16)}`,
            backgroundColor: theme.centerChannelBg,
            color: theme.centerChannelColor,
            fontSize: 12,
            padding: '0 4px',
        },
        dateInput: {
            height: 28,
            borderRadius: 4,
            border: `1px solid ${changeOpacity(theme.centerChannelColor, 0.16)}`,
            backgroundColor: theme.centerChannelBg,
            color: theme.centerChannelColor,
            fontSize: 12,
            padding: '0 4px',
        },
        badgeContainer: {
            display: 'flex',
            gap: 8,
            margin: '4px 0',
            alignItems: 'center',
        },
        priorityHigh: {
            backgroundColor: '#D24B4E',
            color: '#FFFFFF',
            fontSize: 10,
            padding: '2px 6px',
            borderRadius: 10,
            fontWeight: 600,
            textTransform: 'uppercase',
        },
        priorityMedium: {
            backgroundColor: '#FFBC1F',
            color: '#000000',
            fontSize: 10,
            padding: '2px 6px',
            borderRadius: 10,
            fontWeight: 600,
            textTransform: 'uppercase',
        },
        dueDate: {
            fontSize: 11,
            color: changeOpacity(theme.centerChannelColor, 0.64),
            display: 'flex',
            alignItems: 'center',
            fontWeight: 500,
        },
        commentToggle: {
            fontSize: 11,
            color: theme.buttonBg,
            cursor: 'pointer',
            display: 'flex',
            alignItems: 'center',
            marginTop: 4,
            fontWeight: 600,
        },
        commentsSection: {
            marginTop: 8,
            paddingLeft: 42,
            borderTop: `1px solid ${changeOpacity(theme.centerChannelColor, 0.08)}`,
            paddingTop: 8,
        },
        commentsList: {
            marginBottom: 8,
        },
        commentItem: {
            marginBottom: 8,
            fontSize: 12,
        },
        commentHeader: {
            display: 'flex',
            justifyContent: 'space-between',
            marginBottom: 2,
        },
        commentUser: {
            fontWeight: 600,
            color: theme.centerChannelColor,
        },
        commentDate: {
            fontSize: 10,
            color: changeOpacity(theme.centerChannelColor, 0.48),
        },
        commentMessage: {
            color: changeOpacity(theme.centerChannelColor, 0.88),
            whiteSpace: 'pre-wrap',
        },
        noComments: {
            fontSize: 11,
            fontStyle: 'italic',
            color: changeOpacity(theme.centerChannelColor, 0.56),
            marginBottom: 8,
        },
        addCommentContainer: {
            display: 'flex',
            flexDirection: 'column',
            alignItems: 'flex-end',
        },
        commentInput: {
            width: '100%',
            fontSize: 12,
            padding: '8px',
            borderRadius: 4,
            border: `1px solid ${changeOpacity(theme.centerChannelColor, 0.16)}`,
            backgroundColor: theme.centerChannelBg,
            color: theme.centerChannelColor,
            resize: 'none',
            transition: 'border-color 0.2s ease',
        },
        commentDeleteIcon: {
            fontSize: 14,
            marginLeft: 8,
            cursor: 'pointer',
            color: changeOpacity(theme.centerChannelColor, 0.32),
            transition: 'color 0.2s ease',
        },
    };
});

TodoItem.propTypes = {
    remove: PropTypes.func.isRequired,
    issue: PropTypes.object.isRequired,
    theme: PropTypes.object.isRequired,
    siteURL: PropTypes.string.isRequired,
    complete: PropTypes.func.isRequired,
    accept: PropTypes.func.isRequired,
    bump: PropTypes.func.isRequired,
    list: PropTypes.string.isRequired,
    editIssue: PropTypes.func.isRequired,
    openAssigneeModal: PropTypes.func.isRequired,
    setEditingTodo: PropTypes.func.isRequired,
    openTodoToast: PropTypes.func.isRequired,
};

export default TodoItem;
