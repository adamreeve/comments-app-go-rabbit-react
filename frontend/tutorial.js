var converter = new Showdown.converter();

var CommentBox = React.createClass({
  getInitialState: function() {
    return {data: []};
  },

  addComment: function(comment) {
    console.log("Got new comment from server")
    console.log(comment)
    this.setState({data: this.state.data.concat([comment])});
  },

  handleCommentSubmit: function(comment) {
    // Optimistically update before getting update from server
    var comments = this.state.data;
    var newComments = comments.concat([comment]);
    this.setState({data: newComments});

    console.log("Sending new comment over websocket")
    this.webSocket.send(JSON.stringify(comment));
  },

  componentDidMount: function() {
    this.webSocket = new WebSocket(this.props.url)
    this.webSocket.onmessage = (function(e) {
      this.addComment(JSON.parse(e.data));
    }).bind(this)
  },

  render: function() {
    return (
      <div className="commentBox">
        <h1>Comments</h1>
        <CommentList data={this.state.data} />
        <CommentForm onCommentSubmit={this.handleCommentSubmit} />
      </div>
    );
  }
});

var Comment = React.createClass({
  render: function() {
    var rawMarkup = converter.makeHtml(this.props.children.toString())
    return (
      <div className="comment">
        <h2 className="commentAuthor">{this.props.author}</h2>
        <div dangerouslySetInnerHTML={{__html: rawMarkup}} />
      </div>
    );
  }
});

var CommentList = React.createClass({
  render: function() {
    var commentNodes = this.props.data.map(function (comment) {
      return (
          <Comment author={comment.author} key={comment.id}>
            {comment.text}
          </Comment>
        );
    });
    return (
      <div className="commentList">
        {commentNodes}
      </div>
    );
  }
});

var CommentForm = React.createClass({
  handleSubmit: function(e) {
    e.preventDefault();
    var author = this.refs.author.getDOMNode().value.trim();
    var text = this.refs.text.getDOMNode().value.trim();
    if (!text || !author) {
      return;
    }
    var id = guid();
    this.props.onCommentSubmit({author: author, text: text, id: id});
    this.refs.author.getDOMNode().value = '';
    this.refs.text.getDOMNode().value = '';
    return;
  },

  render: function() {
    return (
      <form className="commentForm" onSubmit={this.handleSubmit}>
        <input type="text" placeholder="Your name" ref="author" />
        <input type="text" placeholder="Say something..." ref="text" />
        <input type="submit" value="Post" />
      </form>
    );
  }
});

React.render(
  <CommentBox url="ws://localhost:8080/comments" />,
  document.getElementById("content")
);

function guid(){
  var d = new Date().getTime();
  var uuid = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
    var r = (d + Math.random()*16)%16 | 0;
    d = Math.floor(d/16);
    return (c=='x' ? r : (r&0x3|0x8)).toString(16);
  });
  return uuid;
};
