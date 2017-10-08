﻿namespace goflywin
{
    partial class formMain
    {
        /// <summary>
        /// 必需的设计器变量。
        /// </summary>
        private System.ComponentModel.IContainer components = null;

        /// <summary>
        /// 清理所有正在使用的资源。
        /// </summary>
        /// <param name="disposing">如果应释放托管资源，为 true；否则为 false。</param>
        protected override void Dispose(bool disposing)
        {
            if (disposing && (components != null))
            {
                components.Dispose();
            }
            base.Dispose(disposing);
        }

        #region Windows 窗体设计器生成的代码

        /// <summary>
        /// 设计器支持所需的方法 - 不要修改
        /// 使用代码编辑器修改此方法的内容。
        /// </summary>
        private void InitializeComponent()
        {
            this.labelServer = new System.Windows.Forms.Label();
            this.labelKey = new System.Windows.Forms.Label();
            this.textKey = new System.Windows.Forms.TextBox();
            this.groupAuth = new System.Windows.Forms.GroupBox();
            this.textAuthPass = new System.Windows.Forms.TextBox();
            this.labelAuthPass = new System.Windows.Forms.Label();
            this.textAuthUser = new System.Windows.Forms.TextBox();
            this.labelAuthUser = new System.Windows.Forms.Label();
            this.comboServer = new System.Windows.Forms.ComboBox();
            this.checkPartial = new System.Windows.Forms.CheckBox();
            this.labelUDP = new System.Windows.Forms.Label();
            this.labelUDP_TCP = new System.Windows.Forms.Label();
            this.buttonStart = new System.Windows.Forms.Button();
            this.buttonStop = new System.Windows.Forms.Button();
            this.buttonQuit = new System.Windows.Forms.Button();
            this.buttonDelServer = new System.Windows.Forms.Button();
            this.groupBox1 = new System.Windows.Forms.GroupBox();
            this.listLog = new System.Windows.Forms.ListBox();
            this.labelPort = new System.Windows.Forms.Label();
            this.textPort = new System.Windows.Forms.TextBox();
            this.textUDP_TCP = new System.Windows.Forms.NumericUpDown();
            this.textUDP = new System.Windows.Forms.NumericUpDown();
            this.label1 = new System.Windows.Forms.Label();
            this.comboLogLevel = new System.Windows.Forms.ComboBox();
            this.comboProxyType = new System.Windows.Forms.ComboBox();
            this.label2 = new System.Windows.Forms.Label();
            this.checkAutoMin = new System.Windows.Forms.CheckBox();
            this.labelState = new System.Windows.Forms.Label();
            this.label3 = new System.Windows.Forms.Label();
            this.textDNS = new System.Windows.Forms.NumericUpDown();
            this.groupAuth.SuspendLayout();
            this.groupBox1.SuspendLayout();
            ((System.ComponentModel.ISupportInitialize)(this.textUDP_TCP)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.textUDP)).BeginInit();
            ((System.ComponentModel.ISupportInitialize)(this.textDNS)).BeginInit();
            this.SuspendLayout();
            // 
            // labelServer
            // 
            this.labelServer.AutoSize = true;
            this.labelServer.Location = new System.Drawing.Point(8, 9);
            this.labelServer.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelServer.Name = "labelServer";
            this.labelServer.Size = new System.Drawing.Size(149, 12);
            this.labelServer.TabIndex = 0;
            this.labelServer.Text = "Server Address (IP:Port)";
            // 
            // labelKey
            // 
            this.labelKey.AutoSize = true;
            this.labelKey.Location = new System.Drawing.Point(8, 86);
            this.labelKey.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelKey.Name = "labelKey";
            this.labelKey.Size = new System.Drawing.Size(23, 12);
            this.labelKey.TabIndex = 2;
            this.labelKey.Text = "Key";
            // 
            // textKey
            // 
            this.textKey.Location = new System.Drawing.Point(96, 83);
            this.textKey.Margin = new System.Windows.Forms.Padding(2);
            this.textKey.Name = "textKey";
            this.textKey.Size = new System.Drawing.Size(194, 21);
            this.textKey.TabIndex = 3;
            this.textKey.Text = "0123456789abcdef";
            // 
            // groupAuth
            // 
            this.groupAuth.Controls.Add(this.textAuthPass);
            this.groupAuth.Controls.Add(this.labelAuthPass);
            this.groupAuth.Controls.Add(this.textAuthUser);
            this.groupAuth.Controls.Add(this.labelAuthUser);
            this.groupAuth.Location = new System.Drawing.Point(10, 134);
            this.groupAuth.Margin = new System.Windows.Forms.Padding(2);
            this.groupAuth.Name = "groupAuth";
            this.groupAuth.Padding = new System.Windows.Forms.Padding(2);
            this.groupAuth.Size = new System.Drawing.Size(279, 73);
            this.groupAuth.TabIndex = 4;
            this.groupAuth.TabStop = false;
            this.groupAuth.Text = "User Authentication";
            // 
            // textAuthPass
            // 
            this.textAuthPass.Location = new System.Drawing.Point(86, 44);
            this.textAuthPass.Margin = new System.Windows.Forms.Padding(2);
            this.textAuthPass.Name = "textAuthPass";
            this.textAuthPass.Size = new System.Drawing.Size(190, 21);
            this.textAuthPass.TabIndex = 8;
            // 
            // labelAuthPass
            // 
            this.labelAuthPass.AutoSize = true;
            this.labelAuthPass.Location = new System.Drawing.Point(4, 46);
            this.labelAuthPass.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelAuthPass.Name = "labelAuthPass";
            this.labelAuthPass.Size = new System.Drawing.Size(53, 12);
            this.labelAuthPass.TabIndex = 7;
            this.labelAuthPass.Text = "Password";
            // 
            // textAuthUser
            // 
            this.textAuthUser.Location = new System.Drawing.Point(86, 19);
            this.textAuthUser.Margin = new System.Windows.Forms.Padding(2);
            this.textAuthUser.Name = "textAuthUser";
            this.textAuthUser.Size = new System.Drawing.Size(190, 21);
            this.textAuthUser.TabIndex = 6;
            // 
            // labelAuthUser
            // 
            this.labelAuthUser.AutoSize = true;
            this.labelAuthUser.Location = new System.Drawing.Point(4, 22);
            this.labelAuthUser.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelAuthUser.Name = "labelAuthUser";
            this.labelAuthUser.Size = new System.Drawing.Size(53, 12);
            this.labelAuthUser.TabIndex = 5;
            this.labelAuthUser.Text = "Username";
            // 
            // comboServer
            // 
            this.comboServer.FormattingEnabled = true;
            this.comboServer.Location = new System.Drawing.Point(10, 26);
            this.comboServer.Margin = new System.Windows.Forms.Padding(2);
            this.comboServer.Name = "comboServer";
            this.comboServer.Size = new System.Drawing.Size(220, 20);
            this.comboServer.TabIndex = 5;
            this.comboServer.SelectedIndexChanged += new System.EventHandler(this.comboServer_SelectedIndexChanged);
            // 
            // checkPartial
            // 
            this.checkPartial.AutoSize = true;
            this.checkPartial.Location = new System.Drawing.Point(10, 212);
            this.checkPartial.Margin = new System.Windows.Forms.Padding(2);
            this.checkPartial.Name = "checkPartial";
            this.checkPartial.Size = new System.Drawing.Size(132, 16);
            this.checkPartial.TabIndex = 7;
            this.checkPartial.Text = "Partial encryption";
            this.checkPartial.UseVisualStyleBackColor = true;
            // 
            // labelUDP
            // 
            this.labelUDP.AutoSize = true;
            this.labelUDP.Location = new System.Drawing.Point(8, 235);
            this.labelUDP.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelUDP.Name = "labelUDP";
            this.labelUDP.Size = new System.Drawing.Size(53, 12);
            this.labelUDP.TabIndex = 8;
            this.labelUDP.Text = "UDP Port";
            // 
            // labelUDP_TCP
            // 
            this.labelUDP_TCP.AutoSize = true;
            this.labelUDP_TCP.Location = new System.Drawing.Point(8, 260);
            this.labelUDP_TCP.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelUDP_TCP.Name = "labelUDP_TCP";
            this.labelUDP_TCP.Size = new System.Drawing.Size(47, 12);
            this.labelUDP_TCP.TabIndex = 10;
            this.labelUDP_TCP.Text = "UDP-TCP";
            // 
            // buttonStart
            // 
            this.buttonStart.Location = new System.Drawing.Point(10, 357);
            this.buttonStart.Margin = new System.Windows.Forms.Padding(2);
            this.buttonStart.Name = "buttonStart";
            this.buttonStart.Size = new System.Drawing.Size(90, 26);
            this.buttonStart.TabIndex = 13;
            this.buttonStart.Text = "Start";
            this.buttonStart.UseVisualStyleBackColor = true;
            this.buttonStart.Click += new System.EventHandler(this.buttonStart_Click);
            // 
            // buttonStop
            // 
            this.buttonStop.Enabled = false;
            this.buttonStop.Location = new System.Drawing.Point(105, 357);
            this.buttonStop.Margin = new System.Windows.Forms.Padding(2);
            this.buttonStop.Name = "buttonStop";
            this.buttonStop.Size = new System.Drawing.Size(90, 26);
            this.buttonStop.TabIndex = 14;
            this.buttonStop.Text = "Stop";
            this.buttonStop.UseVisualStyleBackColor = true;
            this.buttonStop.Click += new System.EventHandler(this.buttonStop_Click);
            // 
            // buttonQuit
            // 
            this.buttonQuit.Location = new System.Drawing.Point(199, 357);
            this.buttonQuit.Margin = new System.Windows.Forms.Padding(2);
            this.buttonQuit.Name = "buttonQuit";
            this.buttonQuit.Size = new System.Drawing.Size(90, 26);
            this.buttonQuit.TabIndex = 15;
            this.buttonQuit.Text = "Quit";
            this.buttonQuit.UseVisualStyleBackColor = true;
            this.buttonQuit.Click += new System.EventHandler(this.buttonQuit_Click);
            // 
            // buttonDelServer
            // 
            this.buttonDelServer.Location = new System.Drawing.Point(234, 22);
            this.buttonDelServer.Margin = new System.Windows.Forms.Padding(2);
            this.buttonDelServer.Name = "buttonDelServer";
            this.buttonDelServer.Size = new System.Drawing.Size(56, 26);
            this.buttonDelServer.TabIndex = 17;
            this.buttonDelServer.Text = "Delete";
            this.buttonDelServer.UseVisualStyleBackColor = true;
            this.buttonDelServer.Click += new System.EventHandler(this.buttonDelServer_Click);
            // 
            // groupBox1
            // 
            this.groupBox1.Controls.Add(this.listLog);
            this.groupBox1.Location = new System.Drawing.Point(294, 56);
            this.groupBox1.Margin = new System.Windows.Forms.Padding(2);
            this.groupBox1.Name = "groupBox1";
            this.groupBox1.Padding = new System.Windows.Forms.Padding(2);
            this.groupBox1.Size = new System.Drawing.Size(279, 327);
            this.groupBox1.TabIndex = 18;
            this.groupBox1.TabStop = false;
            this.groupBox1.Text = "Sever Log";
            // 
            // listLog
            // 
            this.listLog.FormattingEnabled = true;
            this.listLog.HorizontalScrollbar = true;
            this.listLog.ItemHeight = 12;
            this.listLog.Location = new System.Drawing.Point(4, 17);
            this.listLog.Margin = new System.Windows.Forms.Padding(2);
            this.listLog.Name = "listLog";
            this.listLog.Size = new System.Drawing.Size(271, 304);
            this.listLog.TabIndex = 0;
            // 
            // labelPort
            // 
            this.labelPort.AutoSize = true;
            this.labelPort.Location = new System.Drawing.Point(8, 58);
            this.labelPort.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelPort.Name = "labelPort";
            this.labelPort.Size = new System.Drawing.Size(83, 12);
            this.labelPort.TabIndex = 19;
            this.labelPort.Text = "Local Listen ";
            // 
            // textPort
            // 
            this.textPort.Location = new System.Drawing.Point(96, 56);
            this.textPort.Margin = new System.Windows.Forms.Padding(2);
            this.textPort.Name = "textPort";
            this.textPort.Size = new System.Drawing.Size(194, 21);
            this.textPort.TabIndex = 20;
            this.textPort.Text = ":8100";
            // 
            // textUDP_TCP
            // 
            this.textUDP_TCP.Location = new System.Drawing.Point(96, 258);
            this.textUDP_TCP.Name = "textUDP_TCP";
            this.textUDP_TCP.Size = new System.Drawing.Size(194, 21);
            this.textUDP_TCP.TabIndex = 21;
            this.textUDP_TCP.Value = new decimal(new int[] {
            3,
            0,
            0,
            0});
            // 
            // textUDP
            // 
            this.textUDP.Location = new System.Drawing.Point(96, 232);
            this.textUDP.Maximum = new decimal(new int[] {
            65535,
            0,
            0,
            0});
            this.textUDP.Minimum = new decimal(new int[] {
            1,
            0,
            0,
            0});
            this.textUDP.Name = "textUDP";
            this.textUDP.Size = new System.Drawing.Size(194, 21);
            this.textUDP.TabIndex = 22;
            this.textUDP.Value = new decimal(new int[] {
            8731,
            0,
            0,
            0});
            // 
            // label1
            // 
            this.label1.AutoSize = true;
            this.label1.Location = new System.Drawing.Point(8, 288);
            this.label1.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.label1.Name = "label1";
            this.label1.Size = new System.Drawing.Size(59, 12);
            this.label1.TabIndex = 23;
            this.label1.Text = "Log Level";
            // 
            // comboLogLevel
            // 
            this.comboLogLevel.DropDownStyle = System.Windows.Forms.ComboBoxStyle.DropDownList;
            this.comboLogLevel.FormattingEnabled = true;
            this.comboLogLevel.Items.AddRange(new object[] {
            "dbg",
            "log",
            "warn",
            "err",
            "off"});
            this.comboLogLevel.Location = new System.Drawing.Point(96, 285);
            this.comboLogLevel.Name = "comboLogLevel";
            this.comboLogLevel.Size = new System.Drawing.Size(193, 20);
            this.comboLogLevel.TabIndex = 24;
            // 
            // comboProxyType
            // 
            this.comboProxyType.DropDownStyle = System.Windows.Forms.ComboBoxStyle.DropDownList;
            this.comboProxyType.FormattingEnabled = true;
            this.comboProxyType.Items.AddRange(new object[] {
            "iplist",
            "global",
            "none"});
            this.comboProxyType.Location = new System.Drawing.Point(96, 109);
            this.comboProxyType.Name = "comboProxyType";
            this.comboProxyType.Size = new System.Drawing.Size(193, 20);
            this.comboProxyType.TabIndex = 26;
            // 
            // label2
            // 
            this.label2.AutoSize = true;
            this.label2.Location = new System.Drawing.Point(8, 113);
            this.label2.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.label2.Name = "label2";
            this.label2.Size = new System.Drawing.Size(65, 12);
            this.label2.TabIndex = 25;
            this.label2.Text = "Proxy Type";
            // 
            // checkAutoMin
            // 
            this.checkAutoMin.AutoSize = true;
            this.checkAutoMin.Checked = true;
            this.checkAutoMin.CheckState = System.Windows.Forms.CheckState.Checked;
            this.checkAutoMin.Location = new System.Drawing.Point(10, 337);
            this.checkAutoMin.Margin = new System.Windows.Forms.Padding(2);
            this.checkAutoMin.Name = "checkAutoMin";
            this.checkAutoMin.Size = new System.Drawing.Size(252, 16);
            this.checkAutoMin.TabIndex = 27;
            this.checkAutoMin.Text = "Minimize to systray when proxy started";
            this.checkAutoMin.UseVisualStyleBackColor = true;
            // 
            // labelState
            // 
            this.labelState.Font = new System.Drawing.Font("Consolas", 18F, System.Drawing.FontStyle.Regular, System.Drawing.GraphicsUnit.Point, ((byte)(0)));
            this.labelState.Location = new System.Drawing.Point(296, 9);
            this.labelState.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.labelState.Name = "labelState";
            this.labelState.Size = new System.Drawing.Size(274, 37);
            this.labelState.TabIndex = 28;
            this.labelState.Text = "NOT RUNNING";
            this.labelState.TextAlign = System.Drawing.ContentAlignment.MiddleCenter;
            this.labelState.Click += new System.EventHandler(this.labelState_Click);
            // 
            // label3
            // 
            this.label3.AutoSize = true;
            this.label3.Location = new System.Drawing.Point(7, 313);
            this.label3.Margin = new System.Windows.Forms.Padding(2, 0, 2, 0);
            this.label3.Name = "label3";
            this.label3.Size = new System.Drawing.Size(59, 12);
            this.label3.TabIndex = 29;
            this.label3.Text = "DNS Cache";
            // 
            // textDNS
            // 
            this.textDNS.Location = new System.Drawing.Point(96, 311);
            this.textDNS.Maximum = new decimal(new int[] {
            65535,
            0,
            0,
            0});
            this.textDNS.Minimum = new decimal(new int[] {
            1,
            0,
            0,
            0});
            this.textDNS.Name = "textDNS";
            this.textDNS.Size = new System.Drawing.Size(194, 21);
            this.textDNS.TabIndex = 30;
            this.textDNS.Value = new decimal(new int[] {
            1024,
            0,
            0,
            0});
            // 
            // formMain
            // 
            this.AutoScaleDimensions = new System.Drawing.SizeF(6F, 12F);
            this.AutoScaleMode = System.Windows.Forms.AutoScaleMode.Font;
            this.ClientSize = new System.Drawing.Size(581, 394);
            this.Controls.Add(this.textDNS);
            this.Controls.Add(this.label3);
            this.Controls.Add(this.labelState);
            this.Controls.Add(this.checkAutoMin);
            this.Controls.Add(this.comboProxyType);
            this.Controls.Add(this.label2);
            this.Controls.Add(this.comboLogLevel);
            this.Controls.Add(this.label1);
            this.Controls.Add(this.textUDP);
            this.Controls.Add(this.textUDP_TCP);
            this.Controls.Add(this.textPort);
            this.Controls.Add(this.labelPort);
            this.Controls.Add(this.groupBox1);
            this.Controls.Add(this.buttonDelServer);
            this.Controls.Add(this.buttonQuit);
            this.Controls.Add(this.buttonStop);
            this.Controls.Add(this.buttonStart);
            this.Controls.Add(this.labelUDP_TCP);
            this.Controls.Add(this.labelUDP);
            this.Controls.Add(this.checkPartial);
            this.Controls.Add(this.comboServer);
            this.Controls.Add(this.groupAuth);
            this.Controls.Add(this.textKey);
            this.Controls.Add(this.labelKey);
            this.Controls.Add(this.labelServer);
            this.FormBorderStyle = System.Windows.Forms.FormBorderStyle.FixedDialog;
            this.Margin = new System.Windows.Forms.Padding(2);
            this.MaximizeBox = false;
            this.Name = "formMain";
            this.ShowIcon = false;
            this.StartPosition = System.Windows.Forms.FormStartPosition.CenterScreen;
            this.Text = "goflywin";
            this.FormClosing += new System.Windows.Forms.FormClosingEventHandler(this.formMain_FormClosing);
            this.Load += new System.EventHandler(this.Form1_Load);
            this.Resize += new System.EventHandler(this.formMain_Resize);
            this.groupAuth.ResumeLayout(false);
            this.groupAuth.PerformLayout();
            this.groupBox1.ResumeLayout(false);
            ((System.ComponentModel.ISupportInitialize)(this.textUDP_TCP)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.textUDP)).EndInit();
            ((System.ComponentModel.ISupportInitialize)(this.textDNS)).EndInit();
            this.ResumeLayout(false);
            this.PerformLayout();

        }

        #endregion

        private System.Windows.Forms.Label labelServer;
        private System.Windows.Forms.Label labelKey;
        private System.Windows.Forms.GroupBox groupAuth;
        private System.Windows.Forms.Label labelAuthPass;
        private System.Windows.Forms.Label labelAuthUser;
        private System.Windows.Forms.Label labelUDP;
        private System.Windows.Forms.Label labelUDP_TCP;
        private System.Windows.Forms.Button buttonStart;
        private System.Windows.Forms.Button buttonStop;
        private System.Windows.Forms.Button buttonQuit;
        private System.Windows.Forms.Button buttonDelServer;
        private System.Windows.Forms.GroupBox groupBox1;
        private System.Windows.Forms.ListBox listLog;
        private System.Windows.Forms.Label labelPort;
        private System.Windows.Forms.Label label1;
        private System.Windows.Forms.Label label2;
        private System.Windows.Forms.CheckBox checkAutoMin;
        private System.Windows.Forms.Label labelState;
        public System.Windows.Forms.TextBox textKey;
        public System.Windows.Forms.TextBox textAuthPass;
        public System.Windows.Forms.TextBox textAuthUser;
        public System.Windows.Forms.ComboBox comboServer;
        public System.Windows.Forms.CheckBox checkPartial;
        public System.Windows.Forms.TextBox textPort;
        public System.Windows.Forms.NumericUpDown textUDP_TCP;
        public System.Windows.Forms.NumericUpDown textUDP;
        public System.Windows.Forms.ComboBox comboLogLevel;
        public System.Windows.Forms.ComboBox comboProxyType;
        private System.Windows.Forms.Label label3;
        public System.Windows.Forms.NumericUpDown textDNS;
    }
}

