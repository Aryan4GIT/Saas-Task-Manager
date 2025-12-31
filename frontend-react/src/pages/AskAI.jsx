import { useState } from 'react';
import { useAuth } from '../context/AuthContext';
import { ragService } from '../services/api.service';
import { Send, Bot, User, Loader2, Sparkles, FileText, AlertCircle, CheckSquare, RefreshCw } from 'lucide-react';
import toast from 'react-hot-toast';

export default function AskAI() {
  const { user } = useAuth();
  const [question, setQuestion] = useState('');
  const [conversation, setConversation] = useState([]);
  const [loading, setLoading] = useState(false);
  const [backfillLoading, setBackfillLoading] = useState(false);

  const isAdmin = user?.role === 'admin';

  const handleBackfill = async () => {
    if (backfillLoading) return;

    setBackfillLoading(true);
    try {
      const response = await ragService.backfill();
      const { tasks_indexed, issues_indexed, errors } = response.data;
      toast.success(`Indexed ${tasks_indexed} tasks and ${issues_indexed} issues${errors > 0 ? ` (${errors} errors)` : ''}`);
    } catch (error) {
      console.error('[AskAI] Backfill failed:', error);
      const message =
        error.response?.data?.error ||
        error.response?.data?.message ||
        'Failed to index data. Please try again.';
      toast.error(message);
    } finally {
      setBackfillLoading(false);
    }
  };

  const handleSubmit = async (e) => {
    e.preventDefault();
    if (!question.trim() || loading) return;

    const userQuestion = question.trim();
    setQuestion('');

    // Add user message to conversation
    setConversation((prev) => [
      ...prev,
      { role: 'user', content: userQuestion },
    ]);

    setLoading(true);

    try {
      const response = await ragService.query(userQuestion);
      const { answer, sources } = response.data;

      // Add AI response to conversation
      setConversation((prev) => [
        ...prev,
        { role: 'assistant', content: answer, sources: sources || [] },
      ]);
    } catch (error) {
      console.error('[AskAI] Query failed:', error);
      const message =
        error.response?.data?.error ||
        error.response?.data?.message ||
        'Failed to get response. Please try again.';
      toast.error(message);

      // Add error message to conversation
      setConversation((prev) => [
        ...prev,
        { role: 'assistant', content: 'Sorry, I encountered an error processing your question. Please try again.', isError: true },
      ]);
    } finally {
      setLoading(false);
    }
  };

  const getSourceIcon = (sourceType) => {
    switch (sourceType) {
      case 'task':
        return <CheckSquare className="w-4 h-4" />;
      case 'issue':
        return <AlertCircle className="w-4 h-4" />;
      default:
        return <FileText className="w-4 h-4" />;
    }
  };

  const getSourceLabel = (sourceType) => {
    switch (sourceType) {
      case 'task':
        return 'Task';
      case 'issue':
        return 'Issue';
      case 'comment':
        return 'Comment';
      default:
        return 'Document';
    }
  };

  const exampleQuestions = [
    'What are the high priority tasks?',
    'Show me open issues',
    'What tasks are overdue?',
    'Summarize recent activity',
  ];

  return (
    <div className="space-y-6">
      {/* Header */}
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-3">
          <div className="p-3 rounded-xl bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30">
            <Sparkles className="w-6 h-6 text-purple-400" />
          </div>
          <div>
            <h1 className="text-2xl font-bold text-white">Ask AI</h1>
            <p className="text-gray-400 text-sm">
              Ask questions about your tasks, issues, and company data
            </p>
          </div>
        </div>

        {/* Admin: Index Data Button */}
        {isAdmin && (
          <button
            onClick={handleBackfill}
            disabled={backfillLoading}
            className="btn bg-slate-800 hover:bg-slate-700 text-white border border-slate-600 flex items-center gap-2"
          >
            {backfillLoading ? (
              <Loader2 className="w-4 h-4 animate-spin" />
            ) : (
              <RefreshCw className="w-4 h-4" />
            )}
            <span>{backfillLoading ? 'Indexing...' : 'Index Data'}</span>
          </button>
        )}
      </div>

      {/* Chat Container */}
      <div className="card min-h-[500px] flex flex-col">
        {/* Messages Area */}
        <div className="flex-1 overflow-y-auto p-4 space-y-4">
          {conversation.length === 0 ? (
            <div className="h-full flex flex-col items-center justify-center text-center py-12">
              <div className="p-4 rounded-full bg-gradient-to-br from-purple-500/20 to-pink-500/20 border border-purple-500/30 mb-4">
                <Bot className="w-12 h-12 text-purple-400" />
              </div>
              <h3 className="text-lg font-semibold text-white mb-2">
                How can I help you today?
              </h3>
              <p className="text-gray-400 text-sm max-w-md mb-6">
                Ask me anything about your tasks, issues, or company data. I'll search through your records and provide relevant answers.
              </p>

              {/* Example Questions */}
              <div className="flex flex-wrap gap-2 justify-center max-w-lg">
                {exampleQuestions.map((q, idx) => (
                  <button
                    key={idx}
                    onClick={() => setQuestion(q)}
                    className="px-3 py-1.5 text-sm bg-slate-800 hover:bg-slate-700 text-gray-300 rounded-lg border border-slate-700 transition-colors"
                  >
                    {q}
                  </button>
                ))}
              </div>
            </div>
          ) : (
            conversation.map((msg, idx) => (
              <div
                key={idx}
                className={`flex gap-3 ${msg.role === 'user' ? 'justify-end' : 'justify-start'}`}
              >
                {msg.role === 'assistant' && (
                  <div className="flex-shrink-0 p-2 rounded-lg bg-purple-500/20 border border-purple-500/30 h-fit">
                    <Bot className="w-5 h-5 text-purple-400" />
                  </div>
                )}

                <div
                  className={`max-w-[80%] ${
                    msg.role === 'user'
                      ? 'bg-gradient-to-r from-primary-500 to-purple-500 text-white'
                      : msg.isError
                      ? 'bg-red-900/30 border border-red-700 text-red-300'
                      : 'bg-slate-800 border border-slate-700 text-gray-200'
                  } rounded-2xl px-4 py-3`}
                >
                  <p className="whitespace-pre-wrap">{msg.content}</p>

                  {/* Sources */}
                  {msg.sources && msg.sources.length > 0 && (
                    <div className="mt-3 pt-3 border-t border-slate-600">
                      <p className="text-xs text-gray-400 mb-2">Sources:</p>
                      <div className="space-y-2">
                        {msg.sources.map((source, sIdx) => (
                          <div
                            key={sIdx}
                            className="flex items-start gap-2 text-xs bg-slate-900/50 rounded-lg p-2"
                          >
                            <span className="flex-shrink-0 mt-0.5 text-gray-400">
                              {getSourceIcon(source.source_type)}
                            </span>
                            <div className="flex-1 min-w-0">
                              <span className="text-purple-400 font-medium">
                                {getSourceLabel(source.source_type)}
                              </span>
                              <p className="text-gray-400 truncate mt-0.5">
                                {source.content?.substring(0, 100)}
                                {source.content?.length > 100 ? '...' : ''}
                              </p>
                            </div>
                            <span className="text-gray-500 flex-shrink-0">
                              {(source.similarity * 100).toFixed(0)}% match
                            </span>
                          </div>
                        ))}
                      </div>
                    </div>
                  )}
                </div>

                {msg.role === 'user' && (
                  <div className="flex-shrink-0 p-2 rounded-lg bg-primary-500/20 border border-primary-500/30 h-fit">
                    <User className="w-5 h-5 text-primary-400" />
                  </div>
                )}
              </div>
            ))
          )}

          {/* Loading indicator */}
          {loading && (
            <div className="flex gap-3 justify-start">
              <div className="flex-shrink-0 p-2 rounded-lg bg-purple-500/20 border border-purple-500/30 h-fit">
                <Bot className="w-5 h-5 text-purple-400" />
              </div>
              <div className="bg-slate-800 border border-slate-700 rounded-2xl px-4 py-3">
                <div className="flex items-center gap-2 text-gray-400">
                  <Loader2 className="w-4 h-4 animate-spin" />
                  <span>Thinking...</span>
                </div>
              </div>
            </div>
          )}
        </div>

        {/* Input Area */}
        <div className="border-t border-slate-700 p-4">
          <form onSubmit={handleSubmit} className="flex gap-3">
            <input
              type="text"
              value={question}
              onChange={(e) => setQuestion(e.target.value)}
              placeholder="Ask a question about your data..."
              className="flex-1 bg-slate-800 border border-slate-700 rounded-xl px-4 py-3 text-white placeholder-gray-500 focus:outline-none focus:ring-2 focus:ring-purple-500 focus:border-transparent"
              disabled={loading}
            />
            <button
              type="submit"
              disabled={!question.trim() || loading}
              className="btn bg-gradient-to-r from-purple-500 to-pink-500 hover:from-purple-600 hover:to-pink-600 text-white px-6 py-3 rounded-xl disabled:opacity-50 disabled:cursor-not-allowed flex items-center gap-2"
            >
              {loading ? (
                <Loader2 className="w-5 h-5 animate-spin" />
              ) : (
                <Send className="w-5 h-5" />
              )}
              <span className="hidden sm:inline">Send</span>
            </button>
          </form>
        </div>
      </div>
    </div>
  );
}
