import React, { useState } from 'react';
import './Chatbot.css';
import { Button } from 'react-bootstrap';

const Chatbot = () => {
  const [isOpen, setIsOpen] = useState(false);
  const [messages, setMessages] = useState([
    { sender: 'bot', text: 'Hello! How can I help you with the exercise schedule today? Type "help" to see what I can do!' }
  ]);
  const [inputValue, setInputValue] = useState('');

  const toggleChat = () => setIsOpen(!isOpen);

  // Format message to handle markdown-style formatting
  const formatMessage = (text) => {
    // Convert markdown bold to HTML
    let formatted = text.replace(/\*\*(.*?)\*\*/g, '<strong>$1</strong>');
    // Convert bullet points
    formatted = formatted.replace(/^• /gm, '&bull; ');
    // Convert newlines to <br>
    formatted = formatted.replace(/\n/g, '<br>');
    // Convert emoji codes if needed
    return formatted;
  };

  const sendMessageToBackend = async (message) => {
    try {
      const response = await fetch('/api/chatbot', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ message }),
      });

      if (!response.ok) {
        throw new Error(`HTTP error! status: ${response.status}`);
      }

      const data = await response.json();
      return data.reply;
    } catch (error) {
      console.error('Error sending message to backend:', error);
      return 'Sorry, I am having trouble connecting to the server.';
    }
  };

  const handleSendMessage = async (e) => {
    e.preventDefault();
    if (inputValue.trim() === '') return;

    const userMessage = { sender: 'user', text: inputValue };
    setMessages(prevMessages => [...prevMessages, userMessage]);

    const botReply = await sendMessageToBackend(inputValue);
    const botResponse = { sender: 'bot', text: botReply };
    setMessages(prevMessages => [...prevMessages, botResponse]);

    setInputValue('');
  };

  return (
    <div className="chatbot-container">
      {isOpen ? (
        <div className="chat-window">
          <div className="chat-header">AOC Assistant</div>
          <div className="chat-messages">
            {messages.map((msg, index) => (
              <div key={index} className={`chat-message ${msg.sender}`}>
                <div dangerouslySetInnerHTML={{ __html: formatMessage(msg.text) }} />
              </div>
            ))}
          </div>
          <form className="chat-input-form" onSubmit={handleSendMessage}>
            <input
              type="text"
              value={inputValue}
              onChange={(e) => setInputValue(e.target.value)}
              placeholder="Type a message..."
            />
            <Button type="submit" variant="primary">Send</Button>
          </form>
        </div>
      ) : (
        <div className="chat-toggle-button" onClick={toggleChat}>
          <span>?</span>
        </div>
      )}
    </div>
  );
};

export default Chatbot;
