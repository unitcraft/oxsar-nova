//jquery.iterator.js
/*
 * jQuery iterator - iterator plugin - v 0.0.2
 * Copyright (c) 2009 ConstNW
 * Licensed under the MIT License:
 * http://www.opensource.org/licenses/mit-license.php
 */
/*
 * Моддификация под Oxsar
 * ICQ: 365-888
 * Forum: http://devtm.ru
*/
(function($){
	$.fn.iterator = function(options){
		var version = '0.0.2';
		var opts = $.extend({}, $.fn.iterator.defaults, options);
		return this.each(function(){
			$this = $(this);
			$this.timerID = null;
			$this.running = false;
			$this.increment = null;
			var o = $.meta ? $.extend({}, opts, $this.data()) : opts;
			$this.startNum = o.startNum;
			$this.stopNum = o.stopNum;
			$this.step = o.step;
			$this.timeout = o.timeout;
			$.fn.iterator.start($this);
		});
	};
	$.fn.iterator.start = function(el){
		$.fn.iterator.stop(el);
		$.fn.iterator.display(el);
	}
	$.fn.iterator.stop = function(el){
		if (el.running) clearTimeout(el.timerID);
		el.running = false;
	}
	$.fn.iterator.display = function(el){
		var value = $.fn.iterator.getValue(el);
		value = el.step >= 0 ? Math.min(value, el.stopNum) : Math.max(value, el.stopNum);
		var re = /([\d{1,3}]+)(\d{3})/;
		el.html(value.toString().replace(re, '$1.$2').replace(re, '$1.$2').replace(re, '$1.$2').replace(re, '$1.$2'));
		if(value == el.stopNum) $.fn.iterator.stop(el);
		el.timerID = setTimeout(function(){ $.fn.iterator.display(el); }, el.timeout);
	}
	$.fn.iterator.getValue = function(el){
		if( el.increment == null )
			el.increment = el.startNum;
		else
			el.increment += el.step;
		return Math.round(el.increment);
	};
	$.fn.iterator.defaults = {
		startNum: 0,
		stopNum: 0,
		step: 0,
		timeout: 1000 // 1000 = one second, 60000 = one minute
	};
})(jQuery);