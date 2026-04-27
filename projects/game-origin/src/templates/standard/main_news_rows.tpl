  {if[0 && (time() / (60*30)) % 2]}
  <tr>
    <td rowspan="2"><nobr><span class="false2"><b>МЕГА АКЦИЯ</b></span></nobr></td>
    <td colspan="2">
      <a href="/forums/index.php?showtopic=1073" target="_blank">
        <span class="true" style="font-weight: bold"><nobr>+27,3% кредитов бесплатно</nobr></span> за SMS
        (оператор a1). Подробнее...
      </a>
    </td>
  </tr>
  <tr>
    <td colspan="2">
      <span class="true" style="font-weight: bold"><nobr>100 кредитов = 0.1 WMZ</nobr></span> - пополни кредиты за так
    </td>
  </tr>
  {/if}

  {if[0]}
  <tr>
    <td><span class="false2"><b>НОВОГОДНЯЯ АКЦИЯ</b></span></td>
    <td colspan="2">
     {if[1]}
      <a href="/forums/index.php?showtopic=1229" target="_blank">
        <b><font color='#00FF00'>+30% </font><font color='#e9b639'>кредитов бесплатно, только в новогодние праздники!  Подробнее...</font></b>
      </a>
     {else}
      <a href="/forums/index.php?showtopic=1229&view=findpost&p=8319" target="_blank">
        <b><font color='#00FF00'>+50% </font><font color='#e9b639'>кредитов бесплатно, только 31 декабря и 1 января! Подробнее...</font></b>
      </a>
     {/if}
    </td>
  </tr>
  {/if}

  {if[0 && OXSAR_RELEASED]}
  <tr>
    <td rowspan="1"><span class="true"><b>КОНКУРС</b></span></td>
    <td colspan="2">
      <a href="/forums/index.php?showtopic=1128" target="_blank">
        Предложи картинки под Ремонтный док, победителю <span class="true">1000 кредитов за каждую картинку</span>.
      </a>
    </td>
  </tr>
  {/if}
  
  {if[0]}
  <tr>
    <td rowspan="1"><span class="true"><b>НОВОСТИ</b></span></td>
    <td colspan="2">
      Идет обновление игры на новую версию. Возможно игра будет недоступна несколько минут. Когда обновление завершится, это сообщение пропадет.
    </td>
  </tr>
  {/if}

  {if[0 && OXSAR_RELEASED]}
  <tr>
    <td rowspan="4"><span class="true"><b>НОВОСТИ</b></span></td>
    <td colspan="2">
      {if[0]}
      <a href="Payment"><font color='#e9b639'><nobr>100 кредитов = 0.1 дол</nobr></font> 
      - пополни кредиты за так, поддержи игру (sms, wm, rb)</a>
      {else if[0]}
      <a href="Payment"><font color='#e9b639'><nobr>+30% кредитов бесплатно</nobr></font> <font color='#ff3300'>(действует временно)</font> пополни кредиты - поддержи игру (SMS, WebMoney, Яндекс.Деньги, телебанк BTБ24, VISA и мн. др.)</a>
      {else}
      <a href="Payment"><font color='#e9b639'><nobr>+50% кредитов бесплатно</nobr></font> - новая акция <font color='#ff3300'>(действует временно)</font> пополни кредиты - поддержи игру (SMS, WebMoney, Яндекс.Деньги, телебанк BTБ24, VISA и мн. др.)</a>
      {/if}
    </td>
  </tr>

  <tr>
    <td colspan="2">
      {if[1]}
      <a href="/forums/index.php?showtopic=1490" target="_blank">
        <span class="true">Версия игры обновлена на Oxsar 1.0.7</span> Подробнее...
      </a>
      {/if}
    </td>
  </tr>

  {if[0]}
  <tr>
    <td colspan="2">
      <a href="/forums/index.php?showtopic=1360" target="_blank">
        Делимся эмоциями о проведенных экспедициях...
      </a>
    </td>
  </tr>
  {/if}

  <tr>
    <td colspan="2">
      Улучшение тех или иных характеристик (количество кораблей, уровни построек и др.) нечестными способами (использование багов, 
      скриптов, мультов и т.п.) влечет бан. 
    </td>
  </tr>

  <tr>
    <td colspan="2">
      {if[((time() / (60*1)) % 6) == 0]}
        Уничтожение Луны выключено. Планируйте это при строительстве флота и атаках.
      {else if[((time() / (60*1)) % 6) == 1]}
        <a href="/forums/index.php?showtopic=1506" target="_blank"><span class="true">Анонс:</span> новый отчет будет с музыкой.</a>
      {else if[((time() / (60*1)) % 6) == 2]}
        <span class="true">Улучшайте Торговое представительство, чтобы ставить на биржу больше лотов для продажи.</span>
      {else if[((time() / (60*1)) % 6) == 3]}
        <a href="/forums/index.php?showtopic=1495" target="_blank"><span class="true">Защита новичка:</span> ликвидировать порог 100.000 очков и ввести лимит <i>10 раз</i>?</a>
      {else if[((time() / (60*1)) % 6) == 4]}
        <a href="/forums/index.php?showtopic=1386" target="_blank">Режим отпуска автоматически выключается через месяц отсутствия игрока в игре.</a>
      {else}
        Выработка ресурсов выключается после 3 суток отсутствия игрока в игре.
      {/if}
    </td>
  </tr>
  {/if}
  
{if[0]}
 Oxsar https://oxsar-nova.ru
 Copyright (c) 2026 oxsar-nova authors. PolyForm Noncommercial 1.0.0.
{/if}