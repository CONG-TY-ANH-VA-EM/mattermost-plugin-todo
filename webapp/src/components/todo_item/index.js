// Copyright (c) 2015-present Mattermost, Inc. All Rights Reserved.
// See LICENSE.txt for license information.

import { connect } from 'react-redux';
import { bindActionCreators } from 'redux';

import { openAssigneeModal, openTodoToast, setEditingTodo, editIssue, fetchComments, addComment, deleteComment } from '../../actions';

import TodoItem from './todo_item';

const mapStateToProps = (state) => {
    return {
        currentUserId: state.entities.users.currentUserId,
    };
};

const mapDispatchToProps = (dispatch) => bindActionCreators({
    editIssue,
    openAssigneeModal,
    setEditingTodo,
    openTodoToast,
    fetchComments,
    addComment,
    deleteComment,
}, dispatch);

export default connect(mapStateToProps, mapDispatchToProps)(TodoItem);
