from flask import Flask, request, jsonify
import time
import random

app = Flask(__name__)

@app.route('/pay', methods=['POST'])
def pay():
    data = request.json
    #chamge this mode to test different scenarios
    mode = data.get('mode', 'random')
    
    if mode == 'timeout':
        time.sleep(10)  # Simulate timeout
        return jsonify({"status": "timeout"}), 504
    elif mode == 'error':
        return jsonify({"status": "error"}), 500
    elif mode == 'declined':
        return jsonify({"status": "declined"}), 400
    elif mode == 'latency':
        time.sleep(random.uniform(1, 3))
        return jsonify({"status": "approved"}), 200
    elif mode == 'random':
        outcomes = ['approved', 'declined', 'error', 'timeout']
        weights = [0.6, 0.15, 0.15, 0.1] # Adjust probabilities as needed
        status = random.choices(outcomes, weights=weights, k=1)[0]
        if status == 'timeout':
            time.sleep(10)
            return jsonify({"status": status}), 504
        elif status == 'error':
            return jsonify({"status": status}), 500
        elif status == 'declined':
            return jsonify({"status": status}), 400
        else:
            return jsonify({"status": status}), 200
    else:  # happy
        return jsonify({"status": "approved"}), 200

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=8080)